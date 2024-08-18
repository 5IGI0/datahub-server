package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/gorilla/mux"
)

func ApiGetDiscourseInstanceState(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	hash_id := mux.Vars(r)["hash_id"]
	var response struct {
		LastTopicId          int64            `json:"last_topic_id"`
		TopicsHighestNumbers map[string]int64 `json:"topics_highest_numbers"`
		FullTopics           []int64          `json:"full_topics"`
		FullUsers            []int64          `json:"full_users"`
	}
	response.TopicsHighestNumbers = make(map[string]int64)
	response.FullTopics = make([]int64, 0)
	response.FullUsers = make([]int64, 0)
	response.LastTopicId = -10
	instance_id := int64(0)

	GlobalContext.Database.Get(&instance_id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", hash_id)
	if instance_id == 0 {
		return response, 404, "", nil
	}

	rows, err := GlobalContext.Database.Queryx("SELECT `topic_id`, `is_data_full` FROM `discourse_topics` WHERE `instance_id`=?", instance_id)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var row DiscourseTopicRow
		AssertError(rows.StructScan(&row))

		if row.TopicId > response.LastTopicId {
			response.LastTopicId = row.TopicId
		}

		if row.IsDataFull != 0 {
			response.FullTopics = append(response.FullTopics, row.TopicId)
		}
	}

	rows, err = GlobalContext.Database.Queryx("SELECT `topic_id`, MAX(`post_id`) AS `id` FROM `discourse_posts` WHERE `instance_id`=? GROUP BY `topic_id`", instance_id)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var row DiscourseTopicRow
		AssertError(rows.StructScan(&row))
		response.TopicsHighestNumbers[fmt.Sprint(row.TopicId)] = row.Id
	}

	rows, err = GlobalContext.Database.Queryx("SELECT `user_id`, `is_data_full` FROM `discourse_users` WHERE `instance_id`=?", instance_id)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var row DiscourseUserRow
		AssertError(rows.StructScan(&row))
		if row.IsDataFull != 0 {
			response.FullUsers = append(response.FullUsers, row.UserId)
		}
	}

	return response, 200, "", nil
}

func ApiAddDiscourseInstance(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var input struct {
		Secure    int8           `json:"secure"`
		Host      string         `json:"host"`
		Root      string         `json:"root"`
		BasicInfo map[string]any `json:"basic_info"`
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var row DiscourseInstanceRow

	row.Secure = input.Secure
	row.Host = input.Host
	row.Root = input.Root
	row.HashId = row.CompHashId()
	row.Title, _ = input.BasicInfo["title"].(string)
	row.Description, _ = input.BasicInfo["description"].(string)
	login_required, _ := input.BasicInfo["login_required"].(bool)
	row.LoginRequired = Ternary(login_required, int8(1), 0)

	row.Title = TruncateText(row.Title, 128)
	row.Description = TruncateText(row.Description, 255)

	raw_data, _ := json.Marshal(input.BasicInfo)
	row.RawData = string(raw_data)

	GlobalContext.Database.Get(&row.Id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", row.HashId)

	if row.Id != 0 {
		q, v := squirrel.Update("discourse_instances").
			SetMap(map[string]any{
				"title":          row.Title,
				"description":    row.Description,
				"login_required": row.LoginRequired,
				"raw_data":       row.RawData,
			}).Where(squirrel.Eq{"id": row.Id}).MustSql()
		GlobalContext.Database.MustExec(q, v...)
	} else {
		q, v := squirrel.Insert("discourse_instances").
			SetMap(map[string]any{
				"secure":         row.Secure,
				"host":           row.Host,
				"root":           row.Root,
				"title":          row.Title,
				"description":    row.Description,
				"login_required": row.LoginRequired,
				"raw_data":       row.RawData,
			}).MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	return map[string]any{
		"hash_id": row.HashId,
	}, 200, "", nil
}

func ApiAddDiscourseCategories(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	hash_id := mux.Vars(r)["hash_id"]
	var input []map[string]any

	instance_id := int64(0)

	GlobalContext.Database.Get(&instance_id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", hash_id)
	if instance_id == 0 {
		return nil, 404, "INSTANCE_NOT_FOUND", errors.New("instance not found")
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var categories []DiscourseCategoryRow
	for _, raw_category := range input {
		var category_row DiscourseCategoryRow
		var e bool

		category_row.InstanceId = instance_id
		category_row.CategoryId, e = ForceInt64Cast(raw_category["id"])
		Assert(e)
		category_row.IsActive = 1
		category_row.Name, _ = raw_category["name"].(string)
		category_row.Slug, _ = raw_category["slug"].(string)
		category_row.Description, _ = raw_category["description"].(string)
		if v, e := raw_category["parent_category_id"]; e {
			category_row.ParentCategoryId.Int64, e = ForceInt64Cast(v)
			Assert(e)
		}
		json_str, err := json.Marshal(raw_category)
		AssertError(err)

		category_row.Name = TruncateText(category_row.Name, 127)
		category_row.Slug = TruncateText(category_row.Slug, 127)
		category_row.Description = TruncateText(category_row.Description, 127)

		category_row.RawData = string(json_str)
		category_row.HashId = category_row.CompHashId()

		categories = append(categories, category_row)
	}

	row2setmap := func(r DiscourseCategoryRow) map[string]any {
		return map[string]any{
			"instance_id":        r.InstanceId,
			"category_id":        r.CategoryId,
			"is_active":          1,
			"name":               r.Name,
			"slug":               r.Slug,
			"description":        r.Description,
			"parent_category_id": r.ParentCategoryId,
			"raw_data":           r.RawData}
	}

	InsertHashIdBasedRows(categories, "discourse_categories", squirrel.Eq{"instance_id": instance_id},
		row2setmap, func(r DiscourseCategoryRow, _ int64) map[string]any { return row2setmap(r) })

	return "ok", 200, "", nil
}

func ApiAddDiscourseTopics(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	hash_id := mux.Vars(r)["hash_id"]
	var input []struct {
		IsFull bool           `json:"is_full"`
		Data   map[string]any `json:"data"`
	}

	instance_id := int64(0)
	GlobalContext.Database.Get(&instance_id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", hash_id)
	if instance_id == 0 {
		return nil, 404, "INSTANCE_NOT_FOUND", errors.New("instance not found")
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var topics []DiscourseTopicRow
	for _, raw_topic := range input {
		var topic DiscourseTopicRow

		topic.InstanceId = instance_id
		topic.TopicId, _ = ForceInt64Cast(raw_topic.Data["id"])
		topic.Title, _ = raw_topic.Data["title"].(string)
		topic.CategoryId, _ = ForceInt64Cast(raw_topic.Data["category_id"])
		topic.UserId.Int64, topic.UserId.Valid = ForceInt64Cast(raw_topic.Data["user_id"])
		topic.IsDataFull = Ternary[int8](raw_topic.IsFull, 1, 0)

		json_str, _ := json.Marshal(raw_topic.Data)
		topic.RawData = string(json_str)
		topic.Title = TruncateText(topic.Title, 128)
		topic.HashId = topic.CompHashId()

		if tag_list, e := raw_topic.Data["tags"].([]any); e {
			var distags []DiscourseTagRow

			tag_descriptions, _ := raw_topic.Data["tags_descriptions"].(map[string]any)
			for _, v := range tag_list {
				var row DiscourseTagRow
				row.Name, _ = v.(string)

				if tag_descriptions != nil {
					row.Description.String, row.Description.Valid = tag_descriptions[row.Name].(string)
				}

				row.HashId = row.CompHashId()
				distags = append(distags, row)
			}

			if len(distags) != 0 {
				__insertDiscourseTags(distags)
			}
		}

		topics = append(topics, topic)
	}

	row2setmap := func(r DiscourseTopicRow) map[string]any {
		return map[string]any{
			"instance_id":  r.InstanceId,
			"topic_id":     r.TopicId,
			"title":        r.Title,
			"category_id":  r.CategoryId,
			"user_id":      r.UserId,
			"raw_data":     r.RawData,
			"is_data_full": r.IsDataFull}
	}

	InsertHashIdBasedRows(topics, "discourse_topics", nil,
		row2setmap,
		func(r DiscourseTopicRow, _ int64) map[string]any {
			return row2setmap(r)
		})

	return "ok", 200, "", nil
}

func ApiAddDiscoursePosts(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	hash_id := mux.Vars(r)["hash_id"]
	var input []map[string]any

	instance_id := int64(0)
	GlobalContext.Database.Get(&instance_id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", hash_id)
	if instance_id == 0 {
		return nil, 404, "INSTANCE_NOT_FOUND", errors.New("instance not found")
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var posts []DiscoursePostRow
	for _, raw_post := range input {
		var post DiscoursePostRow

		post.InstanceId = instance_id
		post.TopicId, _ = ForceInt64Cast(raw_post["topic_id"])
		post.PostId, _ = ForceInt64Cast(raw_post["id"])
		post.UserId, _ = ForceInt64Cast(raw_post["user_id"])

		json_str, _ := json.Marshal(raw_post)
		post.RawData = string(json_str)
		post.HashId = post.CompHashId()

		posts = append(posts, post)
	}

	row2setmap := func(r DiscoursePostRow) map[string]any {
		return map[string]any{
			"instance_id": r.InstanceId,
			"topic_id":    r.TopicId,
			"user_id":     r.UserId,
			"post_id":     r.PostId,
			"raw_data":    r.RawData}
	}

	InsertHashIdBasedRows(posts, "discourse_posts", nil,
		row2setmap,
		func(r DiscoursePostRow, _ int64) map[string]any {
			return row2setmap(r)
		})

	return "ok", 200, "", nil
}

func ApiAddDiscourseUsers(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	hash_id := mux.Vars(r)["hash_id"]
	var input []struct {
		IsFull bool           `json:"is_full"`
		Data   map[string]any `json:"data"`
	}

	instance_id := int64(0)
	GlobalContext.Database.Get(&instance_id, "SELECT `id` FROM `discourse_instances` WHERE hash_id=?", hash_id)
	if instance_id == 0 {
		return nil, 404, "INSTANCE_NOT_FOUND", errors.New("instance not found")
	}

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	var users []DiscourseUserRow
	for _, raw_user := range input {
		var user DiscourseUserRow

		user.InstanceId = instance_id
		user.UserId, _ = ForceInt64Cast(raw_user.Data["id"])
		user.Username, _ = raw_user.Data["username"].(string)
		user.Name, _ = raw_user.Data["name"].(string)
		user.Title, _ = raw_user.Data["title"].(string)

		if website, e := raw_user.Data["website"].(string); e {
			// i assume that the scheme is mendatory
			parts := strings.Split(website, "/")
			if len(parts) >= 3 {
				user.WebSiteDomain.String = strings.Split(parts[2], ":")[0]
				user.WebSiteDomain.Valid = true
			}
		}

		if is_admin, _ := raw_user.Data["admin"].(bool); is_admin {
			user.Flags |= DISCOURSE_USER_ADMIN_FLAG
		}

		if is_mod, _ := raw_user.Data["moderator"].(bool); is_mod {
			user.Flags |= DISCOURSE_USER_MODERATOR_FLAG
		}

		json_str, _ := json.Marshal(raw_user.Data)
		user.RawData = string(json_str)
		user.IsDataFull = Ternary[int8](raw_user.IsFull, 1, 0)

		user.Username = TruncateText(user.Username, 32)
		user.Name = TruncateText(user.Name, 64)
		user.Title = TruncateText(user.Title, 128)
		user.WebSiteDomain.String = TruncateText(user.WebSiteDomain.String, 255)

		user.HashId = user.CompHashId()

		users = append(users, user)
	}

	row2setmap := func(r DiscourseUserRow) map[string]any {
		return map[string]any{
			"instance_id":    r.InstanceId,
			"user_id":        r.UserId,
			"username":       r.Username,
			"name":           r.Name,
			"title":          r.Title,
			"flags":          r.Flags,
			"website_domain": r.WebSiteDomain,
			"raw_data":       r.RawData,
			"is_data_full":   r.IsDataFull}
	}

	InsertHashIdBasedRows(users, "discourse_users", nil,
		row2setmap,
		func(r DiscourseUserRow, _ int64) map[string]any {
			return row2setmap(r)
		})

	return "ok", 202, "", nil
}

func __insertDiscourseTags(Tags []DiscourseTagRow) {
	for i := 0; i < len(Tags); i++ {
		Tags[i].HashId = Tags[i].CompHashId()
	}

	row2setmap := func(r DiscourseTagRow) map[string]any {
		return map[string]any{
			"instance_id": r.InstanceId,
			"name":        r.Name,
			"description": r.Description}
	}

	InsertHashIdBasedRows(Tags, "discourse_tags", nil,
		row2setmap, func(r DiscourseTagRow, _ int64) map[string]any {
			return row2setmap(r)
		})
}

func ApiGetDiscourseInstanceList(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	rows, err := GlobalContext.Database.Queryx(
		"SELECT host, secure, root FROM discourse_instances")
	AssertError(err)

	var ret []string
	for rows.Next() {
		var row DiscourseInstanceRow
		AssertError(rows.StructScan(&row))

		ret = append(ret,
			Ternary(row.Secure == 0, "http", "https")+"://"+row.Host+"/"+row.Root)
	}

	return ret, 200, "", nil
}
