package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"net/url"
	"stash-vr/internal/api/heatmap"
	"stash-vr/internal/config"
	"stash-vr/internal/efile"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/title"
)

type videoData struct {
	Access int `json:"access"`

	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ThumbnailImage string   `json:"thumbnailImage"`
	ThumbnailVideo string   `json:"thumbnailVideo"`
	DateReleased   string   `json:"dateReleased"`
	DateAdded      string   `json:"dateAdded"`
	Duration       float64  `json:"duration"`
	Rating         float32  `json:"rating"`
	Favorites      int      `json:"favorites"`
	IsFavorite     bool     `json:"isFavorite"`
	Projection     string   `json:"projection"`
	Stereo         string   `json:"stereo"`
	Fov            float32  `json:"fov"`
	Lens           string   `json:"lens"`
	EventServer    string   `json:"eventServer"`
	Scripts        []script `json:"scripts"`
	Tags           []tag    `json:"tags"`
	Media          []media  `json:"media"`

	WriteFavorite bool `json:"writeFavorite"`
	WriteRating   bool `json:"writeRating"`
	WriteTags     bool `json:"writeTags"`
}

type media struct {
	Name    string   `json:"name,omitempty"`
	Sources []source `json:"sources"`
}

type source struct {
	Resolution int    `json:"resolution"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Size       int    `json:"size"`
	Url        string `json:"url"`
}

type script struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, baseUrl string, videoId string, includeMediaSource bool) (videoData, error) {
	sceneId, eFileSuffix, isEScene := efile.GetSceneIdAndEFileSuffix(videoId)

	findSceneResponse, err := gql.FindSceneFull(ctx, client, sceneId)
	if err != nil {
		return videoData{}, fmt.Errorf("FindSceneFull: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return videoData{}, fmt.Errorf("FindSceneFull: not found")
	}
	s := findSceneResponse.FindScene.SceneFullParts

	if len(s.SceneScanParts.Files) == 0 {
		return videoData{}, fmt.Errorf("scene %s has no files", sceneId)
	}

	thumbnailUrl := stash.ApiKeyed(s.Paths.Screenshot)
	if !config.Get().IsHeatmapDisabled && s.ScriptParts.Interactive && s.ScriptParts.Paths.Interactive_heatmap != "" {
		thumbnailUrl = heatmap.GetCoverUrl(baseUrl, sceneId)
	}

	title := title.GetSceneTitle(s.Title, s.GetFiles()[0].Basename)

	var eventServer string

	if isEScene && config.Get().EventServerUrl != "" {
		eventServerUrl, _ := url.Parse(config.Get().EventServerUrl + "/events")
		eventServerQuery, _ := url.ParseQuery(eventServerUrl.RawQuery)
		eventServerQuery.Add("filename", s.GetFiles()[0].Basename)
		eventServerQuery.Add("esuffix", eFileSuffix)
		eventServerUrl.RawQuery = eventServerQuery.Encode()
		eventServer = eventServerUrl.String()
		log.Ctx(ctx).Debug().Str("eventServer", eventServer).Msg("Providing eventServer")
		title = efile.MakeESceneTitleWithEFileSuffix(title, eFileSuffix)
	}

	vd := videoData{
		Access:         1,
		Title:          title,
		Description:    s.Details,
		ThumbnailImage: thumbnailUrl,
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		DateReleased:   s.Date,
		DateAdded:      s.Created_at.Format("2006-01-02"),
		Duration:       s.SceneScanParts.Files[0].Duration * 1000,
		Rating:         float32(s.Rating100) / 20,
		Favorites:      s.O_counter,
		WriteFavorite:  true,
		WriteRating:    true,
		WriteTags:      true,
		EventServer:    eventServer,
	}

	setIsFavorite(s, &vd)

	if includeMediaSource {
		setMediaSources(ctx, s, &vd)
	} else {
		vd.Media = []media{{Sources: []source{{}}}}
	}

	set3DFormat(s, &vd)

	setTags(s, &vd)

	setScripts(s, &vd)

	return vd, nil
}

func setTags(s gql.SceneFullParts, videoData *videoData) {
	tags := getTags(s.SceneScanParts)
	videoData.Tags = tags
}

func setScripts(s gql.SceneFullParts, videoData *videoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, script{
			Name: "Script-" + s.Title,
			Url:  stash.ApiKeyed(s.ScriptParts.Paths.Funscript),
		})
	}
}

func set3DFormat(s gql.SceneFullParts, videoData *videoData) {
	for _, t := range s.Tags {
		switch t.Name {
		case "DOME":
			videoData.Projection = "equirectangular"
			videoData.Stereo = "sbs"
			continue
		case "SPHERE":
			videoData.Projection = "equirectangular360"
			videoData.Stereo = "sbs"
			continue
		case "FISHEYE":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			continue
		case "MKX200":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			videoData.Lens = "MKX200"
			videoData.Fov = 200.0
			continue
		case "RF52":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			videoData.Fov = 190.0
			continue
		case "CUBEMAP":
			videoData.Projection = "cubemap"
			videoData.Stereo = "sbs"
		case "EAC":
			videoData.Projection = "equiangularCubemap"
			videoData.Stereo = "sbs"
		case "SBS":
			videoData.Stereo = "sbs"
			continue
		case "TB":
			videoData.Stereo = "tb"
			continue
		}
	}
}

func setMediaSources(ctx context.Context, s gql.SceneFullParts, videoData *videoData) {
	for _, stream := range stash.GetStreams(ctx, s.StreamsParts, true) {
		e := media{
			Name: stream.Name,
		}
		for _, s := range stream.Sources {
			vs := source{
				Resolution: s.Resolution,
				Url:        s.Url,
			}
			e.Sources = append(e.Sources, vs)
		}
		videoData.Media = append(videoData.Media, e)
	}
}

func setIsFavorite(s gql.SceneFullParts, videoData *videoData) {
	videoData.IsFavorite = ContainsFavoriteTag(s.TagPartsArray)
}

func ContainsFavoriteTag(ts gql.TagPartsArray) bool {
	for _, t := range ts.Tags {
		if t.Name == config.Get().FavoriteTag {
			return true
		}
	}
	return false
}
