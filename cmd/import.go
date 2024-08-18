package cmd

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/kr/pretty"
	"github.com/maruel/natural"
	"github.com/nanoteck137/sewaddle-core/library"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

var importCmd = &cobra.Command{
	Use: "import",
	// Run: func(cmd *cobra.Command, args []string) {
	// 	lib, err := library.ReadFromDir("/Volumes/media/manga")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	// lib.AddSerie()
	//
	// 	pretty.Println(lib)
	// },
}

type OldChapter struct {
	Index int      `json:"index"`
	Name  string   `json:"name"`
	Pages []string `json:"pages"`
}

type OldManga struct {
	Title    string       `json:"title"`
	Chapters []OldChapter `json:"chapters"`
}

var importOldCmd = &cobra.Command{
	Use:  "old <DIR>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("args[0]: %v\n", args[0])

		output, _ := cmd.Flags().GetString("output")

		p := args[0]

		mangaInfoPath := path.Join(args[0], "manga.json")

		data, err := os.ReadFile(mangaInfoPath)
		if err != nil {
			log.Fatal(err)
		}

		var info OldManga
		err = json.Unmarshal(data, &info)
		if err != nil {
			log.Fatal(err)
		}

		pretty.Println(info)

		serieSlug := library.Slug(info.Title)

		p = path.Join(output, serieSlug)
		err = os.Mkdir(p, 0755)
		if err != nil {
			log.Fatal(err)
		}

		var chapters []library.ChapterMetadata

		for _, chapter := range info.Chapters {
			slug := library.Slug(chapter.Name)

			chapterPath := path.Join(p, slug)
			err := os.Mkdir(chapterPath, 0755)
			if err != nil {
				log.Fatal(err)
			}

			for _, page := range chapter.Pages {
				pagePath := path.Join(args[0], "chapters", strconv.Itoa(chapter.Index), page)
				_, err := Copy(pagePath, path.Join(chapterPath, page))
				if err != nil {
					log.Fatal(err)
				}
			}

			fmt.Println("Done Copying:", slug)

			chapters = append(chapters, library.ChapterMetadata{
				Slug:  slug,
				Name:  chapter.Name,
				Pages: chapter.Pages,
			})
		}

		serie := library.SerieMetadata{
			Slug:     serieSlug,
			Title:    info.Title,
			CoverArt: "",
			Chapters: chapters,
		}

		// slices.SortFunc(serie.Chapters, func(a, b library.ChapterMetadata) int {
		// 	return strings.Compare(a.Slug, b.Slug)
		// })

		sort.Slice(serie.Chapters, func(i, j int) bool {
			x := serie.Chapters[i]
			y := serie.Chapters[j]

			return natural.Less(x.Slug, y.Slug)
		})

		data, err = toml.Marshal(serie)
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(path.Join(p, "manga.toml"), data, 0644)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Done")
	},
}

var importCbzCmd = &cobra.Command{
	Use:  "cbz <CBZ_FILE>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("args[0]: %v\n", args[0])

		lib, err := library.ReadFromDir("./work")
		if err != nil {
			log.Fatal(err)
		}

		_ = lib

		r, err := zip.OpenReader(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		var comicInfo *zip.File

		for _, f := range r.File {
			if f.Name == "ComicInfo.xml" {
				comicInfo = f
			}
		}

		if comicInfo != nil {
			var pages []string
			for _, f := range r.File {
				ext := path.Ext(f.Name)
				if ext == ".jpeg" || ext == ".jpg" || ext == ".png" {
					pages = append(pages, f.Name)
				}
			}

			fmt.Println("Found ComicInfo.xml")

			reader, err := comicInfo.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			if err != nil {
				log.Fatal(err)
			}

			type ComicInfo struct {
				Title  string `xml:"Title"`
				Series string `xml:"Series"`
			}

			var info ComicInfo
			err = xml.Unmarshal(data, &info)
			if err != nil {
				log.Fatal(err)
			}

			pretty.Println(info)

			chapterSlug := library.Slug(info.Title)
			serieSlug := library.Slug(info.Series)

			fmt.Printf("info.Series: %v\n", info.Series)
			fmt.Printf("library.Slug(info.Series): %v\n", library.Slug(info.Series))
			fmt.Printf("library.Slug(info.Title): %v\n", library.Slug(info.Title))

			var serie library.SerieMetadata
			found := false

			for _, s := range lib.Series {
				if s.Slug == serieSlug {
					serie = s
					found = true
					break
				}
			}

			if found {
				pretty.Println(serie)

				for _, c := range serie.Chapters {
					if c.Slug == chapterSlug {
						log.Fatalf("Chapter already exists: '%s'", chapterSlug)
					}
				}

				serie.Chapters = append(serie.Chapters, library.ChapterMetadata{
					Slug:  chapterSlug,
					Name:  info.Title,
					Pages: pages,
				})

				slices.SortFunc(serie.Chapters, func(a, b library.ChapterMetadata) int {
					return strings.Compare(a.Slug, b.Slug)
				})

				p := serie.Path()

				d, err := toml.Marshal(serie)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("string(d): %v\n", string(d))

				err = os.WriteFile(path.Join(p, "manga.toml"), d, 0644)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				p := path.Join(lib.Base, serieSlug)
				err = os.Mkdir(p, 0755)
				if err != nil {
					if !os.IsExist(err) {
						log.Fatal(err)
					}
				}

				serie := library.SerieMetadata{
					Slug:     serieSlug,
					Title:    info.Series,
					CoverArt: "",
					Chapters: []library.ChapterMetadata{
						{
							Slug:  library.Slug(info.Title),
							Name:  info.Title,
							Pages: pages,
						},
					},
				}

				d, err := toml.Marshal(serie)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("string(d): %v\n", string(d))

				err = os.WriteFile(path.Join(p, "manga.toml"), d, 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	},
}

func init() {
	importOldCmd.Flags().StringP("output", "o", ".", "Output Directory")

	importCmd.AddCommand(importCbzCmd)
	importCmd.AddCommand(importOldCmd)

	rootCmd.AddCommand(importCmd)
}
