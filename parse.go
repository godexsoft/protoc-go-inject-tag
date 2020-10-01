package main

import (
	"fmt"
	"strings"
)

func tagFromComment(comment string) (tag string) {
	match := rTagComment.FindStringSubmatch(comment)
	if len(match) == 2 {
		tag = match[1]
	}
	return
}

func fieldFromComment(comment string) (tag string) {
	match := rFieldComment.FindStringSubmatch(comment)
	if len(match) == 2 {
		tag = match[1]
	}
	return
}

type tagItem struct {
	key   string
	value string
}

type tagItems []tagItem

func (ti tagItems) format() string {
	tags := []string{}
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func (ti tagItems) override(nti tagItems) tagItems {
	overrided := []tagItem{}
	for i := range ti {
		var dup = -1
		for j := range nti {
			if ti[i].key == nti[j].key {
				dup = j
				break
			}
		}
		if dup == -1 {
			overrided = append(overrided, ti[i])
		} else {
			overrided = append(overrided, nti[dup])
			nti = append(nti[:dup], nti[dup+1:]...)
		}
	}
	return append(overrided, nti...)
}

func newTagItems(tag string) tagItems {
	items := []tagItem{}
	splitted := rTags.FindAllString(tag, -1)

	for _, t := range splitted {
		sepPos := strings.Index(t, ":")
		items = append(items, tagItem{
			key:   t[:sepPos],
			value: t[sepPos+1:],
		})
	}
	return items
}

func injectTag(contents []byte, area textArea) (injected []byte) {
	expr := make([]byte, area.End-area.Start)
	copy(expr, contents[area.Start-1:area.End-1])
	cti := newTagItems(area.CurrentTag)
	iti := newTagItems(area.InjectTag)
	ti := cti.override(iti)
	expr = rInject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))
	injected = append(injected, contents[:area.Start-1]...)
	injected = append(injected, expr...)

	if strings.Contains(area.InjectTag, "rel:has-one") {
		field := strings.Split(string(expr), " ")[0] + "ID"
		injected = append(injected, fmt.Sprintf("\n\t// #inject_tag: generated go-pg ID field for 'rel:has-one' \n\t%s %s", field, "int64")...)
	}

	injected = append(injected, contents[area.End-1:]...)

	return
}

func injectField(contents []byte, area textArea) (injected []byte) {
	injected = append(injected, contents[:area.Start-1]...)
	injected = append(injected, fmt.Sprintf("\t// #inject_field: generated go-pg ID primary key\n\t%s\n\n", area.InjectField)...)
	injected = append(injected, contents[area.End-1:]...)

	return
}
