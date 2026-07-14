package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"endfield-music/internal/service"

	"github.com/spf13/cobra"
)

var netease = service.NewNeteaseService("http://127.0.0.1:3000")

var rootCmd = &cobra.Command{
	Use:   "endfield-music-cli",
	Short: "集成音乐系统命令行工具",
	Long:  `集成音乐系统命令行工具，支持搜索、下载歌曲和歌单`,
}

var searchCmd = &cobra.Command{
	Use:   "search [关键词]",
	Short: "搜索歌曲",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := strings.Join(args, " ")
		limit, _ := cmd.Flags().GetInt("limit")

		body, err := netease.SearchSongs(keyword, limit, 0)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resultData, ok := result["result"].(map[string]interface{}); ok {
			if songs, ok := resultData["songs"].([]interface{}); ok {
				fmt.Printf("找到 %v 首歌曲:\n\n", resultData["songCount"])
				for i, s := range songs {
					song := s.(map[string]interface{})
					artist := ""
					if ar, ok := song["artists"].([]interface{}); ok && len(ar) > 0 {
						artist = ar[0].(map[string]interface{})["name"].(string)
					}
					fmt.Printf("%d. %s - %s (ID: %v)\n", i+1, song["name"], artist, song["id"])
				}
			}
		}
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download [歌曲ID]",
	Short: "下载歌曲",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("无效的歌曲ID: %s", args[0])
		}

		quality, _ := cmd.Flags().GetString("quality")
		output, _ := cmd.Flags().GetString("output")

		// 获取歌曲详情
		body, err := netease.GetSongDetail(id)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if songs, ok := result["songs"].([]interface{}); ok && len(songs) > 0 {
			song := songs[0].(map[string]interface{})
			name := song["name"].(string)
			artist := ""
			if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
				artist = ar[0].(map[string]interface{})["name"].(string)
			}

			fmt.Printf("正在下载: %s - %s\n", artist, name)

			// 获取音质对应的 bitrate
			br := 320000
			if quality == "standard" {
				br = 128000
			} else if quality == "lossless" {
				br = 999000
			}

			// 获取歌曲 URL
			urlBody, err := netease.GetSongURL(id, br, "")
			if err != nil {
				return err
			}

			var urlResult map[string]interface{}
			json.Unmarshal(urlBody, &urlResult)

			if data, ok := urlResult["data"].([]interface{}); ok && len(data) > 0 {
				songData := data[0].(map[string]interface{})
				url, _ := songData["url"].(string)
				if url == "" {
					return fmt.Errorf("歌曲无版权或需要 VIP")
				}

				ext := ".mp3"
				if quality == "lossless" {
					ext = ".flac"
				}

				filename := fmt.Sprintf("%s - %s%s", artist, name, ext)
				if output != "" {
					filename = output
				}

				fmt.Printf("保存为: %s\n", filename)
				fmt.Println("下载中...")

				// 下载文件
				resp, err := netease.GetHTTPClient().Get(url)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				out, err := os.Create(filename)
				if err != nil {
					return err
				}
				defer out.Close()

				buf := make([]byte, 32*1024)
				for {
					n, err := resp.Body.Read(buf)
					if n > 0 {
						out.Write(buf[:n])
					}
					if err != nil {
						break
					}
				}

				fmt.Println("下载完成!")
			}
		}

		return nil
	},
}

var playlistCmd = &cobra.Command{
	Use:   "playlist [歌单ID]",
	Short: "下载歌单",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("无效的歌单ID: %s", args[0])
		}

		quality, _ := cmd.Flags().GetString("quality")

		body, err := netease.GetPlaylistDetail(id, "")
		if err != nil {
			return err
		}

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if playlist, ok := result["playlist"].(map[string]interface{}); ok {
			fmt.Printf("歌单: %s\n", playlist["name"])

			if trackIds, ok := playlist["trackIds"].([]interface{}); ok {
				fmt.Printf("共 %d 首曲目\n\n", len(trackIds))

				for i, tid := range trackIds {
					t := tid.(map[string]interface{})
					songID := int(t["id"].(float64))

					fmt.Printf("[%d/%d] 下载歌曲 ID: %d\n", i+1, len(trackIds), songID)

					// 获取歌曲详情
					detailBody, err := netease.GetSongDetail(songID)
					if err != nil {
						fmt.Printf("  获取详情失败: %v\n", err)
						continue
					}

					var detailResult map[string]interface{}
					json.Unmarshal(detailBody, &detailResult)

					if songs, ok := detailResult["songs"].([]interface{}); ok && len(songs) > 0 {
						song := songs[0].(map[string]interface{})
						name := song["name"].(string)
						artist := ""
						if ar, ok := song["ar"].([]interface{}); ok && len(ar) > 0 {
							artist = ar[0].(map[string]interface{})["name"].(string)
						}

						fmt.Printf("  %s - %s\n", artist, name)

						// 获取 URL 并下载
						br := 320000
						if quality == "standard" {
							br = 128000
						} else if quality == "lossless" {
							br = 999000
						}

						urlBody, err := netease.GetSongURL(songID, br, "")
						if err != nil {
							fmt.Printf("  获取 URL 失败: %v\n", err)
							continue
						}

						var urlResult map[string]interface{}
						json.Unmarshal(urlBody, &urlResult)

						if data, ok := urlResult["data"].([]interface{}); ok && len(data) > 0 {
							songData := data[0].(map[string]interface{})
							url, _ := songData["url"].(string)
							if url == "" {
								fmt.Println("  无版权或需要 VIP")
								continue
							}

							ext := ".mp3"
							if quality == "lossless" {
								ext = ".flac"
							}

							filename := fmt.Sprintf("%s - %s%s", artist, name, ext)
							resp, err := netease.GetHTTPClient().Get(url)
							if err != nil {
								fmt.Printf("  下载失败: %v\n", err)
								continue
							}

							out, err := os.Create(filename)
							if err != nil {
								resp.Body.Close()
								fmt.Printf("  创建文件失败: %v\n", err)
								continue
							}

							buf := make([]byte, 32*1024)
							for {
								n, err := resp.Body.Read(buf)
								if n > 0 {
									out.Write(buf[:n])
								}
								if err != nil {
									break
								}
							}

							out.Close()
							resp.Body.Close()
							fmt.Println("  下载完成")
						}
					}
				}
			}
		}

		return nil
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录管理",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("登录管理")
		fmt.Println("\n请使用 Web 界面进行登录: http://localhost:33550")
		fmt.Println("登录后 Cookie 将自动同步到 CLI")
	},
}

func init() {
	searchCmd.Flags().Int("limit", 30, "搜索结果数量")

	downloadCmd.Flags().StringP("quality", "q", "high", "音质: standard, high, lossless")
	downloadCmd.Flags().StringP("output", "o", "", "输出文件名")

	playlistCmd.Flags().StringP("quality", "q", "high", "音质: standard, high, lossless")

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(playlistCmd)
	rootCmd.AddCommand(loginCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
