package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	version "github.com/hashicorp/go-version"

	"github.com/Sirupsen/logrus"
)

var (
	//use map so that we can find if an item is contained in array with O(1)
	groups = make(map[string]map[string]bool)
)

func stripWhitespaces(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func after(datestr string) bool {
	time1, err := time.Parse(time.RFC3339, datestr)
	if err != nil {
		panic(fmt.Errorf("Date %s is invalid. err=%s", datestr, err))
	}
	return time.Now().After(time1)
}

func before(datestr string) bool {
	time1, err := time.Parse(time.RFC3339, datestr)
	if err != nil {
		panic(fmt.Errorf("Date %s is invalid. err=%s", datestr, err))
	}
	return time.Now().Before(time1)
}

func randomPerc(percent int, reference interface{}, seed int) bool {
	return randomPercRange(0, percent, reference, seed)
}

func randomPercRange(fromPercent int, toPercent int, reference interface{}, seed int) bool {
	fromPerc := (fromPercent * 65535) / 100
	toPerc := (toPercent * 65535) / 100
	hashb := md5.Sum([]byte(fmt.Sprintf("%d%s", seed, reference)))
	hv := binary.BigEndian.Uint16(hashb[:8])
	return (hv >= uint16(fromPerc) && hv < uint16(toPerc))
}

func versionCheck(versionStr string, condition string) bool {
	v1, err := version.NewVersion(versionStr)
	if err != nil {
		logrus.Warnf("Error on version check. version='%s'; condition='%s'", versionStr, condition)
		return false
	}
	constraints, err := version.NewConstraint(condition)
	return constraints.Check(v1)
}

func match(value string, regex string) bool {
	re := regexp.MustCompile(regex)
	return re.MatchString(value)
}

func groupContains(groupName string, element string) bool {
	_, exists := groups[groupName][element]
	return exists
}

func loadGroupFromFile(groups map[string]map[string]bool, groupName string, sourceFile string) {
	logrus.Infof("Loading group %s elements from %s into memory", groupName, sourceFile)
	_, exists := groups[groupName]
	if !exists {
		groups[groupName] = make(map[string]bool)
	}
	file, err := os.Open(sourceFile)
	if err != nil {
		logrus.Errorf("Error loading group '%s' elements from file '%s'. err=%s", groupName, sourceFile, err)
		os.Exit(1)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("Error loading group '%s' elements from file '%s'. err=%s", groupName, sourceFile, err)
		os.Exit(1)
	}
	contents := string(b)
	contents = strings.Replace(contents, " ", ",", -1)
	contents = strings.Replace(contents, ";", ",", -1)
	contents = strings.Replace(contents, "\n", ",", -1)
	elements := strings.Split(contents, ",")
	loadGroupArray(groups, groupName, elements)
}

func loadGroupArray(groups map[string]map[string]bool, groupName string, elements []string) {
	logrus.Infof("Loading group %s elements from array into memory", groupName)
	_, exists := groups[groupName]
	if !exists {
		groups[groupName] = make(map[string]bool)
	}
	for _, elem := range elements {
		el := strings.TrimSpace(elem)
		if el != "" {
			groups[groupName][elem] = true
		}
	}
}
