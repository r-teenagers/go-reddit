package reddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

// ModerationService handles communication with the moderation
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_moderation
type ModerationService struct {
	client *Client
}

// ModAction is an action executed by a moderator of a subreddit, such
// as inviting another user to be a mod, or setting permissions.
type ModAction struct {
	ID      string     `json:"id,omitempty"`
	Action  string     `json:"action,omitempty"`
	Created *Timestamp `json:"created_utc,omitempty"`
	Details string     `json:"details,omitempty"`

	Moderator string `json:"mod,omitempty"`
	// Not the full ID, just the ID36.
	ModeratorID string `json:"mod_id36,omitempty"`

	// The author of whatever the action was produced on, e.g. a user, post, comment, etc.
	TargetAuthor string `json:"target_author,omitempty"`
	// This is the full ID of whatever the target was.
	TargetID        string `json:"target_fullname,omitempty"`
	TargetTitle     string `json:"target_title,omitempty"`
	TargetPermalink string `json:"target_permalink,omitempty"`
	TargetBody      string `json:"target_body,omitempty"`

	Subreddit string `json:"subreddit,omitempty"`
	// Not the full ID, just the ID36.
	SubredditID string `json:"sr_id36,omitempty"`
}

// ModPermissions are the different permissions moderators have or don't have on a subreddit.
// Read about them here: https://mods.reddithelp.com/hc/en-us/articles/360009381491-User-Management-moderators-and-permissions
type ModPermissions struct {
	All          bool `permission:"all"`
	Access       bool `permission:"access"`
	ChatConfig   bool `permission:"chat_config"`
	ChatOperator bool `permission:"chat_operator"`
	Config       bool `permission:"config"`
	Flair        bool `permission:"flair"`
	Mail         bool `permission:"mail"`
	Posts        bool `permission:"posts"`
	Wiki         bool `permission:"wiki"`
}

func (p *ModPermissions) String() (s string) {
	if p == nil {
		return "+all"
	}

	t := reflect.TypeOf(*p)
	v := reflect.ValueOf(*p)

	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).Kind() != reflect.Bool {
			continue
		}

		permission := t.Field(i).Tag.Get("permission")
		permitted := v.Field(i).Bool()

		if permitted {
			s += "+"
		} else {
			s += "-"
		}

		s += permission

		if i != t.NumField()-1 {
			s += ","
		}
	}

	return
}

// BanConfig configures the ban of the user being banned.
type BanConfig struct {
	Reason string `url:"reason,omitempty"`
	// Not visible to the user being banned.
	ModNote string `url:"note,omitempty"`
	// How long the ban will last. 0-999. Leave nil for permanent.
	Days *int `url:"duration,omitempty"`
	// Note to include in the ban message to the user.
	Message string `url:"ban_message,omitempty"`
}

// Actions gets a list of moderator actions on a subreddit.
func (s *ModerationService) Actions(ctx context.Context, subreddit string, opts *ListModActionOptions) ([]*ModAction, *Response, error) {
	path := fmt.Sprintf("r/%s/about/log", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, resp, err
	}
	return l.ModActions(), resp, nil
}

// AcceptInvite accepts a pending invite to moderate the specified subreddit.
func (s *ModerationService) AcceptInvite(ctx context.Context, subreddit string) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/accept_moderator_invite", subreddit)

	form := url.Values{}
	form.Set("api_type", "json")

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Approve a post or comment via its full ID.
func (s *ModerationService) Approve(ctx context.Context, id string) (*Response, error) {
	path := "api/approve"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Remove a post, comment or modmail message via its full ID.
func (s *ModerationService) Remove(ctx context.Context, id string) (*Response, error) {
	path := "api/remove"

	form := url.Values{}
	form.Set("id", id)
	form.Set("spam", "false")

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// RemoveSpam removes a post, comment or modmail message via its full ID and marks it as spam.
func (s *ModerationService) RemoveSpam(ctx context.Context, id string) (*Response, error) {
	path := "api/remove"

	form := url.Values{}
	form.Set("id", id)
	form.Set("spam", "true")

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Leave abdicates your moderator status in a subreddit via its full ID.
func (s *ModerationService) Leave(ctx context.Context, subredditID string) (*Response, error) {
	path := "api/leavemoderator"

	form := url.Values{}
	form.Set("id", subredditID)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// LeaveContributor abdicates your approved user status in a subreddit via its full ID.
func (s *ModerationService) LeaveContributor(ctx context.Context, subredditID string) (*Response, error) {
	path := "api/leavecontributor"

	form := url.Values{}
	form.Set("id", subredditID)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Reported returns posts and comments that have been reported.
func (s *ModerationService) Reported(ctx context.Context, subreddit string, opts *ListOptions) ([]*Post, []*Comment, *Response, error) {
	path := fmt.Sprintf("r/%s/about/reports", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, nil, resp, err
	}
	return l.Posts(), l.Comments(), resp, nil
}

// Spam returns posts and comments marked as spam.
func (s *ModerationService) Spam(ctx context.Context, subreddit string, opts *ListOptions) ([]*Post, []*Comment, *Response, error) {
	path := fmt.Sprintf("r/%s/about/spam", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, nil, resp, err
	}
	return l.Posts(), l.Comments(), resp, nil
}

// Queue returns posts and comments requiring moderator reviews, such as one that have been
// reported or caught in the spam filter.
func (s *ModerationService) Queue(ctx context.Context, subreddit string, opts *ListOptions) ([]*Post, []*Comment, *Response, error) {
	path := fmt.Sprintf("r/%s/about/modqueue", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, nil, resp, err
	}
	return l.Posts(), l.Comments(), resp, nil
}

// Unmoderated returns posts that have yet to be approved/removed by a mod.
func (s *ModerationService) Unmoderated(ctx context.Context, subreddit string, opts *ListOptions) ([]*Post, *Response, error) {
	path := fmt.Sprintf("r/%s/about/unmoderated", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, resp, err
	}
	return l.Posts(), resp, nil
}

// Edited gets posts and comments that have been edited recently.
func (s *ModerationService) Edited(ctx context.Context, subreddit string, opts *ListOptions) ([]*Post, []*Comment, *Response, error) {
	path := fmt.Sprintf("r/%s/about/edited", subreddit)
	l, resp, err := s.client.getListing(ctx, path, opts)
	if err != nil {
		return nil, nil, resp, err
	}
	return l.Posts(), l.Comments(), resp, nil
}

// IgnoreReports prevents reports on a post or comment from causing notifications.
func (s *ModerationService) IgnoreReports(ctx context.Context, id string) (*Response, error) {
	path := "api/ignore_reports"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// UnignoreReports allows reports on a post or comment to cause notifications.
func (s *ModerationService) UnignoreReports(ctx context.Context, id string) (*Response, error) {
	path := "api/unignore_reports"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Invite a user to become a moderator of the subreddit.
// If permissions is nil, all permissions will be granted.
func (s *ModerationService) Invite(ctx context.Context, subreddit string, username string, permissions *ModPermissions) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/friend", subreddit)

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", "moderator_invite")
	form.Set("permissions", permissions.String())

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Uninvite a user from becoming a moderator of the subreddit.
func (s *ModerationService) Uninvite(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "moderator_invite")
}

// SetPermissions sets the mod permissions for the user in the subreddit.
// If permissions is nil, all permissions will be granted.
func (s *ModerationService) SetPermissions(ctx context.Context, subreddit string, username string, permissions *ModPermissions) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/setpermissions", subreddit)

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", "moderator_invite")
	form.Set("permissions", permissions.String())

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Ban a user from the subreddit.
func (s *ModerationService) Ban(ctx context.Context, subreddit string, username string, config *BanConfig) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/friend", subreddit)

	form, err := query.Values(config)
	if err != nil {
		return nil, err
	}

	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", "banned")

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Unban a user from the subreddit.
func (s *ModerationService) Unban(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "banned")
}

// BanWiki bans a user from contributing to the subreddit wiki.
func (s *ModerationService) BanWiki(ctx context.Context, subreddit string, username string, config *BanConfig) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/friend", subreddit)

	form, err := query.Values(config)
	if err != nil {
		return nil, err
	}

	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", "wikibanned")

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// UnbanWiki unbans a user from contributing to the subreddit wiki.
func (s *ModerationService) UnbanWiki(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "wikibanned")
}

// Mute a user in the subreddit.
func (s *ModerationService) Mute(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.createRelationship(ctx, subreddit, username, "muted")
}

// Unmute a user in the subreddit.
func (s *ModerationService) Unmute(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "muted")
}

// ApproveUser adds a user as an approved user to the subreddit.
func (s *ModerationService) ApproveUser(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.createRelationship(ctx, subreddit, username, "contributor")
}

// UnapproveUser removes a user as an approved user to the subreddit.
func (s *ModerationService) UnapproveUser(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "contributor")
}

// ApproveUserWiki adds a user as an approved wiki contributor in the subreddit.
func (s *ModerationService) ApproveUserWiki(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.createRelationship(ctx, subreddit, username, "wikicontributor")
}

// UnapproveUserWiki removes a user as an approved wiki contributor in the subreddit.
func (s *ModerationService) UnapproveUserWiki(ctx context.Context, subreddit string, username string) (*Response, error) {
	return s.deleteRelationship(ctx, subreddit, username, "wikicontributor")
}

func (s *ModerationService) createRelationship(ctx context.Context, subreddit, username, relationship string) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/friend", subreddit)

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", relationship)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *ModerationService) deleteRelationship(ctx context.Context, subreddit, username, relationship string) (*Response, error) {
	path := fmt.Sprintf("r/%s/api/unfriend", subreddit)

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("name", username)
	form.Set("type", relationship)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Distinguish your post or comment via its full ID, adding a moderator tag to it.
// todo: add how=admin and how=special? They require special privileges.
func (s *ModerationService) Distinguish(ctx context.Context, id string) (*Response, error) {
	path := "api/distinguish"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("how", "yes")
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// DistinguishAndSticky your comment via its full ID, adding a moderator tag to it
// and stickying the comment at the top of the thread.
func (s *ModerationService) DistinguishAndSticky(ctx context.Context, id string) (*Response, error) {
	path := "api/distinguish"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("how", "yes")
	form.Set("sticky", "true")
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Undistinguish your post or comment via its full ID, removing the moderator tag from it.
func (s *ModerationService) Undistinguish(ctx context.Context, id string) (*Response, error) {
	path := "api/distinguish"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("how", "no")
	form.Set("id", id)

	req, err := s.client.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

type ModmailParticipant struct {
	IsMod         bool   `json:"isMod"`
	IsAdmin       bool   `json:"isAdmin"`
	Name          string `json:"name"`
	IsOp          bool   `json:"isOp"`
	IsParticipant bool   `json:"isParticipant"`
	IsApproved    bool   `json:"isApproved"`
	IsHidden      bool   `json:"isHidden"`
	Id            string `json:"id"`
	IsDeleted     bool   `json:"isDeleted"`
}

type ModmailOwner struct {
	DisplayName string `json:"displayName,omitempty"`
	Type        string `json:"type,omitempty"`
	Id          string `json:"id,omitempty"`
}

type ModmailMessage struct {
	Body            string             `json:"body"`
	Author          ModmailParticipant `json:"author"`
	IsInternal      bool               `json:"isInternal"`
	Date            Timestamp          `json:"date"`
	BodyMarkdown    string             `json:"bodyMarkdown"`
	Id              string             `json:"id"`
	ParticipatingAs string             `json:"participatingAs"`
}

type Modmail struct {
	FullyLoaded bool

	Id                   string                 `json:"id"`
	IsAuto               bool                   `json:"isAuto,omitempty"`
	Participant          ModmailParticipant     `json:"participant,omitempty"`
	IsRepliable          bool                   `json:"isRepliable,omitempty"`
	IsInternal           bool                   `json:"isInternal,omitempty"`
	LastModUpdate        Timestamp              `json:"lastModUpdate,omitempty"`
	Authors              []ModmailParticipant   `json:"authors,omitempty"`
	LastUpdated          Timestamp              `json:"lastUpdated,omitempty"`
	ParticipantSubreddit map[string]interface{} `json:"participantSubreddit,omitempty"`
	State                int                    `json:"state,omitempty"`
	ConversationType     string                 `json:"conversationType,omitempty"`
	LastUnread           Timestamp              `json:"lastUnread,omitempty"`
	Owner                ModmailOwner           `json:"owner,omitempty"`
	Subject              string                 `json:"subject,omitempty"`
	IsHighlighted        bool                   `json:"isHighlighted,omitempty"`
	NumMessages          int                    `json:"numMessages,omitempty"`
	ObjIds               []struct {
		Id  string
		Key string
	}

	Messages []ModmailMessage
}

type ListModmailOptions struct {
	// Maximum number of items to be returned.
	// The default is 25 and max is 100.
	Limit int `url:"limit,omitempty"`

	// The full ID of an item in the listing to use
	// as the anchor point of the list. Only items
	// appearing after it will be returned.
	After string `url:"after,omitempty"`

	// Comma separated list of subreddit names.
	Entity string `url:"entity,omitempty"`

	// One of "recent", "mod", "user", "unread"
	Sort string `url:"sort,omitempty"`

	// One of "all", "appeals", "notifications", "inbox", "filtered", "inprogress", "mod",
	// "archived", "default", "highlighted", "join_requests", "new"
	State string `url:"state,omitempty"`
}

func (m *ModerationService) GetModmail(ctx context.Context, subreddit string, opts ListModmailOptions) ([]Modmail, *Response, error) {
	path, err := addOptions("api/mod/conversations", opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := m.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(struct {
		Conversation map[string]Modmail        `json:"conversations"`
		Messages     map[string]ModmailMessage `json:"messages"`
	})
	_ = root

	resp, err := m.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	conversations := make([]Modmail, 0, len(root.Conversation))
	for _, m := range root.Conversation {
		for _, obj := range m.ObjIds {
			if obj.Key != "messages" {
				continue
			}

			message, ok := root.Messages[obj.Id]
			if !ok {
				continue
			}

			m.Messages = append(m.Messages, message)
		}

		conversations = append(conversations, m)
	}

	return conversations, resp, nil
}

func (m *ModerationService) ReloadModmail(ctx context.Context, modmail *Modmail, markRead bool) error {
	path := fmt.Sprintf("api/mod/conversations/%s", modmail.Id)

	opts := new(struct {
		markRead bool
	})
	opts.markRead = markRead

	path, err := addOptions(path, opts)
	if err != nil {
		return err
	}

	req, err := m.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	root := new(struct {
		Conversation Modmail                   `json:"conversation"`
		Messages     map[string]ModmailMessage `json:"messages"`
	})
	_ = root

	_, err = m.client.Do(ctx, req, root)
	if err != nil {
		return err
	}

	*modmail = root.Conversation
	msgs := make([]ModmailMessage, 0, modmail.NumMessages)

	for _, obj := range modmail.ObjIds {
		if obj.Key != "messages" {
			continue
		}

		message, ok := root.Messages[obj.Id]
		if !ok {
			continue
		}

		msgs = append(msgs, message)
	}

	modmail.Messages = msgs
	modmail.FullyLoaded = true
	return nil
}
