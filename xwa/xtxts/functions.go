package xtxts

import (
	"encoding/json"
	"fmt"

	"github.com/askasoft/pango/cog/hashmap"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pangox/xwa"
)

func GetStrings(locale, name string, defs ...string) []string {
	return str.Fields(tbs.GetText(locale, name, defs...))
}

func GetInts(locale, name string, defs ...string) []int {
	ss := GetStrings(locale, name, defs...)
	ns := make([]int, len(ss))
	for i, s := range ss {
		ns[i] = num.Atoi(s)
	}
	return ns
}

func GetInt64s(locale, name string, defs ...string) []int64 {
	ss := GetStrings(locale, name, defs...)
	ns := make([]int64, len(ss))
	for i, s := range ss {
		ns[i] = num.Atol(s)
	}
	return ns
}

func GetLinkedHashMap(locale, name string) *linkedhashmap.LinkedHashMap[string, string] {
	m := &linkedhashmap.LinkedHashMap[string, string]{}
	if err := m.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(locale, name))); err != nil {
		panic(fmt.Errorf("invalid [%s] %s: %w", locale, name, err))
	}
	return m
}

func GetReverseMap(locale, name string) *hashmap.HashMap[string, string] {
	m := make(map[string]string)
	if err := json.Unmarshal(str.UnsafeBytes(tbs.GetText(locale, name)), &m); err != nil {
		panic(fmt.Errorf("invalid [%s] %s: %w", locale, name, err))
	}

	rm := hashmap.NewHashMap[string, string]()
	for k, v := range m {
		rm.Set(v, k)
	}
	return rm
}

func GetAllReverseMap(name string) *hashmap.HashMap[string, string] {
	am := hashmap.NewHashMap[string, string]()
	for _, lang := range xwa.Locales {
		rm := GetReverseMap(lang, name)
		am.Copy(rm)
	}
	return am
}
