package githubutils

import (
	"fmt"
	"regexp"
	"strings"

	"aztfy-download-counter/database"
)

func ParseTagName(tagName string, contentType string) (version string, osType database.OsType, arch string, err error) {
	switch contentType {
	case "application/zip":
		return ParseTagNameForZip(tagName)
	case "application/x-msdownload":
		return ParseTagNameForMsi(tagName)
	case "application/gzip":
		return ParseTagNameForGz(tagName)
	default:
		return "", "", "", fmt.Errorf("parse failed")
	}
}

func ParseTagNameForZip(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?m).*_(v\d*\.\d*\.\d*)_([a-z]+)_(.+)\.zip`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsType(strings.ToLower(result[2])), result[3], nil
}

func ParseTagNameForMsi(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`.*_(v\d*\.\d*\.\d*)_(.+)\.msi`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 3 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsTypeWindows, result[2], nil
}

func ParseTagNameForGz(tagName string) (version string, osType database.OsType, arch string, err error) {
	reg := regexp.MustCompile(`(?mU).*_(v{0,1}\d*\.\d*\.\d*)_(.+)_(.+)\.tar\.gz`)
	result := reg.FindStringSubmatch(tagName)
	if len(result) != 4 {
		return "", "", "", fmt.Errorf("parse failed")
	}
	return result[1], database.OsType(strings.ToLower(result[2])), result[3], nil
}
