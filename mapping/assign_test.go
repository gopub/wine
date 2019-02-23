package mapping_test

import (
	"github.com/gopub/mapper"
	"testing"
)

type Image struct {
	Width  int    `mapper:"w,min=100,max=800"`
	Height int    `mapper:"h,min=100,max=800"`
	Link   string `mapper:"pattern=url"`
}

type Topic struct {
	Title      string   `mapper:"min=2,max=30"`
	CoverImage *Image   `mapper:"optional"`
	MoreImages []*Image `mapper:"optional"`
}

func TestAssign(t *testing.T) {
	params := map[string]interface{}{
		"title": "this is title",
		"cover_image": map[string]interface{}{
			"w":    100,
			"h":    200,
			"link": "https://www.image.com",
		},
		"more_images": []map[string]interface{}{
			{
				"w":    100,
				"h":    200,
				"link": "https://www.image.com",
			},
		},
	}

	var topic *Topic
	err := mapper.Assign(&topic, params)
	if err != nil {
		t.FailNow()
	}
}

func TestAssignSlice(t *testing.T) {
	params := map[string]interface{}{
		"title": "this is title",
		"cover_image": map[string]interface{}{
			"w":    100,
			"h":    200,
			"link": "https://www.image.com",
		},
		"more_images": []map[string]interface{}{
			{
				"w":    100,
				"h":    200,
				"link": "https://www.image.com",
			},
		},
	}

	values := []interface{}{params}
	var topics []*Topic
	err := mapper.Assign(&topics, values)
	if err != nil || len(topics) == 0 {
		t.FailNow()
	}
}

type User struct {
	Id       int
	Name     string
	OpenAuth *OpenAuth
}

type OpenAuth struct {
	Provider string
	OpenID   string
}

type UserInfo struct {
	Id       int
	Name     string
	OpenAuth *OpenAuthInfo
}

type OpenAuthInfo struct {
	Provider string
	OpenID   string
}

func TestAssignStruct(t *testing.T) {
	user := &User{}
	userInfo := &UserInfo{
		Id:   1,
		Name: "tom",
		OpenAuth: &OpenAuthInfo{
			Provider: "wechat",
			OpenID:   "open_id_123",
		},
	}

	err := mapper.Assign(user, userInfo)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("%#v", user)
}
