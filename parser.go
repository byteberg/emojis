// Copyright © ByteBerg, 2024-01-28
// Author website: https://byteberg.net
// Unicode emoji file parser

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type EmojiEntry struct {
	CodePoints      []string `json:"codePoints"`
	Status          string   `json:"status"`
	EmojiName       string   `json:"emojiName"`
	EmojiGroupID    int      `json:"emojiGroupID"`
	EmojiSubgroupID int      `json:"emojiSubgroupID"`
}

type EmojiGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EmojiSubgroup struct {
	ID      int    `json:"id"`
	GroupID int    `json:"groupID"`
	Name    string `json:"name"`
}

type FinalResult struct {
	Groups    []EmojiGroup    `json:"groups"`
	Subgroups []EmojiSubgroup `json:"subGroups"`
	Emojis    []EmojiEntry    `json:"emojis"`
}

func main() {
	// https://unicode.org/Public/emoji/
	filePath := "emoji-test.txt"
	emojiData, err := readEmojiFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	emojiGroups, emojiSubgroups, emojiEntries := parseEmojiData(emojiData)

	res := FinalResult{
		Groups:    emojiGroups,
		Subgroups: emojiSubgroups,
		Emojis:    emojiEntries,
	}

	//resBytes, err := json.MarshalIndent(res, "", "    ")
	resBytes, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("emojis.json", resBytes, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("emojiGroups:", len(emojiGroups))
	fmt.Println("emojiSubgroups:", len(emojiSubgroups))
	fmt.Println("emojis:", len(emojiEntries))
}

func readEmojiFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func parseEmojiData(emojiData []string) ([]EmojiGroup, []EmojiSubgroup, []EmojiEntry) {
	var emojiEntries []EmojiEntry

	var groups []EmojiGroup
	var subgroups []EmojiSubgroup
	groupID := 0
	subgroupID := 0

	for _, line := range emojiData {
		if strings.HasPrefix(line, "# group:") {
			// Group
			groupName := strings.Replace(line, "# group:", "", -1)
			groupName = strings.TrimSpace(groupName)

			groupID++
			groups = append(groups, EmojiGroup{
				ID:   groupID,
				Name: groupName,
			})

		} else if strings.HasPrefix(line, "# subgroup:") {
			// Subgroup
			subgroupName := strings.Replace(line, "# subgroup:", "", -1)
			subgroupName = strings.TrimSpace(subgroupName)

			subgroupID++
			subgroups = append(subgroups, EmojiSubgroup{
				ID:      subgroupID,
				GroupID: groupID,
				Name:    subgroupName,
			})

		} else if !strings.HasPrefix(line, "#") && strings.Contains(line, ";") {
			// Emoji Line

		}
	}

	var currentGroup EmojiGroup
	var currentSubgroup EmojiSubgroup
	for _, line := range emojiData {
		if strings.HasPrefix(line, "# group:") {
			// Group
			groupName := strings.Replace(line, "# group:", "", -1)
			groupName = strings.TrimSpace(groupName)

			currentSubgroup = EmojiSubgroup{}
			for _, group := range groups {
				if group.Name == groupName {
					currentGroup = group
					break
				}
			}

		} else if strings.HasPrefix(line, "# subgroup:") {
			// Subgroup
			subgroupName := strings.Replace(line, "# subgroup:", "", -1)
			subgroupName = strings.TrimSpace(subgroupName)

			for _, subgroup := range subgroups {
				if subgroup.Name == subgroupName {
					currentSubgroup = subgroup
					break
				}
			}

		} else if !strings.HasPrefix(line, "#") && strings.Contains(line, ";") {
			// Emoji Line

			var codePoints, status, emoji, version, name string

			line = "\"" + line
			line = strings.Replace(line, " ; ", "\" ; ", 1)

			// Format: code points; status # emoji EX.X name
			_, err := fmt.Sscanf(line, "%q ; %s # %s E%s %s", &codePoints, &status, &emoji, &version, &name)
			if err != nil {
				fmt.Println("Error parsing:", err)
				return nil, nil, nil
			}

			codePointList := strings.Fields(codePoints)

			emojiEntries = append(emojiEntries, EmojiEntry{
				CodePoints:      codePointList,
				Status:          status,
				EmojiName:       name,
				EmojiGroupID:    currentGroup.ID,
				EmojiSubgroupID: currentSubgroup.ID,
			})

			//fmt.Println("Code Point:", codePointList)
			//fmt.Println("Status:", status)
			//fmt.Println("Emoji:", emoji)
			//fmt.Println("Version:", version)
			//fmt.Println("Name:", name)
		}
	}

	return groups, subgroups, emojiEntries
}
