package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	assetMeta "language-change/asset-meta"
	excelLanguage "language-change/excel-language"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func isValidLang(lang string) bool {
	valid := []string{"en", "jp", "cn", "kr"}
	return slices.Contains(valid, lang)
}

func GetLanguage(assetPath string) (string, string, error) {
	typeVersionGame := "cn"

	indexHash, err := assetMeta.GetIndexHash(assetPath)
	if err != nil {
		return "", "", err
	}

	DesignIndex, err := assetMeta.DesignIndexFromBytes(assetPath, indexHash)
	if err != nil {
		return "", "", err
	}
	dataEntry, fileEntry, err := DesignIndex.FindDataAndFileByTarget(-515329346)
	if err != nil {
		return "", "", err
	}
	allowedLanguage := excelLanguage.NewExcelLanguage(assetPath, &dataEntry, &fileEntry)
	languageRows, err := allowedLanguage.Parse()
	if err != nil {
		return "", "", err
	}

	currentTextLang := ""
	currentVoiceLang := ""

	pairs := []struct {
		area string
		typ  *uint8
	}{
		{"os", nil},
		{"cn", func() *uint8 { v := uint8(1); return &v }()},
		{"os", func() *uint8 { v := uint8(1); return &v }()},
		{"cn", nil},
	}

	for _, p := range pairs {
		var found *excelLanguage.LanguageRow
		for i := range languageRows {
			if languageRows[i].Area != nil && *languageRows[i].Area == p.area {
				if (languageRows[i].Type == nil && p.typ == nil) ||
					(languageRows[i].Type != nil && p.typ != nil && *languageRows[i].Type == *p.typ) {
					found = &languageRows[i]
					break
				}
			}
		}
		if found == nil {
			continue
		}
		if found.DefaultLanguage != nil && found.Area != nil && *found.Area == typeVersionGame && found.Type == nil {
			currentTextLang = *found.DefaultLanguage
		}
		if found.DefaultLanguage != nil && found.Area != nil && *found.Area == typeVersionGame && found.Type != nil {
			currentVoiceLang = *found.DefaultLanguage
		}
	}

	if currentTextLang == "" || currentVoiceLang == "" || !isValidLang(currentTextLang) || !isValidLang(currentVoiceLang) {
		return "", "", errors.New("not found language")
	}

	return currentTextLang, currentVoiceLang, nil
}

func SetLanguage(assetPath string, text, voice string) error {
	indexHash, err := assetMeta.GetIndexHash(assetPath)
	if err != nil {
		return err
	}

	DesignIndex, err := assetMeta.DesignIndexFromBytes(assetPath, indexHash)
	if err != nil {
		return err
	}

	dataEntry, fileEntry, err := GetAssetData(DesignIndex, "AllowedLanguage")
	if err != nil {
		return err
	}
	allowedLanguage := excelLanguage.NewExcelLanguage(assetPath, dataEntry, fileEntry)
	languageRows, err := allowedLanguage.Parse()
	if err != nil {
		return err
	}

	pairs := []struct {
		area string
		typ  *uint8
		lang string
	}{
		{"os", nil, text},
		{"cn", func() *uint8 { v := uint8(1); return &v }(), voice},
		{"os", func() *uint8 { v := uint8(1); return &v }(), voice},
		{"cn", nil, text},
	}

	for _, p := range pairs {
		var found *excelLanguage.LanguageRow
		for i := range languageRows {
			if languageRows[i].Area != nil && *languageRows[i].Area == p.area {
				if (languageRows[i].Type == nil && p.typ == nil) ||
					(languageRows[i].Type != nil && p.typ != nil && *languageRows[i].Type == *p.typ) {
					found = &languageRows[i]
					break
				}
			}
		}
		if found == nil {
			continue
		}

		found.DefaultLanguage = &p.lang
		found.LanguageList = []string{p.lang}
	}

	data, err := allowedLanguage.Unmarshal(languageRows)
	if err != nil {
		return err
	}

	filePath := filepath.Join(assetPath, fileEntry.FileByteName+".bytes")

	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(int64(dataEntry.Offset), 0); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}

	if len(data) < int(dataEntry.Size) {
		remaining := int(dataEntry.Size) - len(data)
		zeros := bytes.Repeat([]byte{0}, remaining)
		if _, err := f.Write(zeros); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter source DesignData folder path: ")
	source, _ := reader.ReadString('\n')
	source = strings.TrimSpace(source)
	if source == "" {
		fmt.Fprintln(os.Stderr, "no source folder provided")
		os.Exit(1)
	}

	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "source folder does not exist")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error accessing path:", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintln(os.Stderr, "path is not a directory")
		os.Exit(1)
	}


	currentTextLang, currentVoiceLang, err := GetLanguage(source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting language:", err)
		os.Exit(1)
	}

	fmt.Println("Current language:", currentTextLang, currentVoiceLang)

	fmt.Println("Allow languages: en, jp, cn, kr")

	fmt.Print("Enter text language: ")
	textLang, _ := reader.ReadString('\n')
	textLang = strings.TrimSpace(textLang)
	if textLang == "" {
		fmt.Fprintln(os.Stderr, "no text language provided")
		os.Exit(1)
	}

	fmt.Print("Enter voice language: ")
	voiceLang, _ := reader.ReadString('\n')
	voiceLang = strings.TrimSpace(voiceLang)
	if voiceLang == "" {
		fmt.Fprintln(os.Stderr, "no voice language provided")
		os.Exit(1)
	}

	if err := SetLanguage(source, textLang, voiceLang); err != nil {
		fmt.Fprintln(os.Stderr, "error setting language:", err)
		os.Exit(1)
	}

	fmt.Println("Language set successfully.")

}
