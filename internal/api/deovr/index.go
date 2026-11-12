package deovr

import (
	"context"

	"github.com/rs/zerolog/log"
	"stash-vr/internal/library"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
)

type indexDto struct {
	Authorized string     `json:"authorized"`
	Scenes     []sceneDto `json:"scenes"`
}

type sceneDto struct {
	Name string           `json:"name"`
	List []previewDataDto `json:"list"`
}

type previewDataDto struct {
	Id           string  `json:"id"`
	ThumbnailUrl *string `json:"thumbnailUrl"`
	Title        string  `json:"title"`
	VideoLength  int     `json:"videoLength"`
	VideoUrl     string  `json:"video_url"`
}

func buildIndex(ctx context.Context, sections []library.Section, vds map[string]*library.VideoData, baseUrl string) (indexDto, error) {
	index := indexDto{Authorized: "1", Scenes: make([]sceneDto, len(sections))}

	for i, section := range sections {
		s := sceneDto{
			Name: section.Name,
			List: make([]previewDataDto, 0, len(section.Ids)),
		}
		index.Scenes[i] = s

		for _, sectionSceneId := range section.Ids {
			vd := vds[sectionSceneId]
			if vd == nil {
				log.Ctx(ctx).Warn().Str("sceneId", sectionSceneId).Msg("scene not found in library, skipping")
				continue
			}
			if len(vd.SceneParts.Files) == 0 {
				log.Ctx(ctx).Warn().Str("sceneId", vd.SceneParts.Id).Str("title", vd.Title()).Msg("scene has no files, skipping")
				continue
			}
			dto := previewDataDto{
				Id:          vd.SceneParts.Id,
				Title:       vd.Title(),
				VideoLength: int(vd.SceneParts.Files[0].Duration),
				VideoUrl:    getVideoDataUrl(baseUrl, vd.Id()),
			}
			if vd.SceneParts.Paths.Screenshot != nil {
				dto.ThumbnailUrl = util.Ptr(stash.ApiKeyed(*vd.SceneParts.Paths.Screenshot))
			}
			s.List = append(s.List, dto)
		}
	}

	return index, nil
}
