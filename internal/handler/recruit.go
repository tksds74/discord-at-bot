package handler

import (
	"at-bot/internal/discord"
	"at-bot/internal/recruit"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type interactionCustomID string

// ボタンインタラクション識別子
const (
	interactionJoin    interactionCustomID = "recruit/join"
	interactionDecline interactionCustomID = "recruit/decline"
	interactionClose   interactionCustomID = "recruit/close"
	interactionCancel  interactionCustomID = "recruit/cancel"
)

// UI用文字列
const (
	joinLabel    = "🙋 参加"
	declineLabel = "🙅 不参加"
)

// customID共通キー
const (
	messageIDKey = "messageID"
	customIDKey  = "customID"
)

func (id interactionCustomID) toString() string {
	return string(id)
}

type openRecruitCommand struct {
	service *recruit.RecruitUsecase
}

func NewOpenRecruitCommand(service *recruit.RecruitUsecase) *openRecruitCommand {
	return &openRecruitCommand{
		service: service,
	}
}

func (command *openRecruitCommand) Prefix() string {
	return "@"
}

func (command *openRecruitCommand) Handle(session *discordgo.Session, message *discordgo.MessageCreate) error {
	// 定員引数の取得
	num, err := command.extractArgNumber(message.Content)
	if err != nil {
		// @はメンションなどにも使用されるので数値が来なくてもエラーにはしない
		return nil
	}

	// 初期状態の募集メッセージを作成、送信
	initialState := InitState(message.Author.ID, num)
	sentMessage, err := session.ChannelMessageSendComplex(message.ChannelID, initialState.toMessageContent())
	if err != nil {
		return fmt.Errorf("failed to send message. channelId: %s, %w", message.ChannelID, err)
	}

	ctx, cancel := createContextWithTimeout()
	defer cancel()

	// 募集の作成
	_, err = command.service.Open(
		ctx,
		recruit.GuildID(message.GuildID),
		recruit.ChannelID(message.ChannelID),
		recruit.MessageID(sentMessage.ID),
		num,
		recruit.UserID(message.Author.ID),
	)

	if err != nil {
		// エラーが発生した場合は送信したメッセージを削除
		_ = session.ChannelMessageDelete(message.ChannelID, sentMessage.ID)
		return err
	}

	// 作成者のコマンドメッセージを削除
	err = session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		// Note: 元メッセージは消せなくてもよいのでログだけ残す
		log.Printf("failed to delete message. channelId: %s, messageId: %s", message.ChannelID, message.ID)
	}

	return nil
}

func (command *openRecruitCommand) extractArgNumber(content string) (int, error) {
	arg := strings.TrimSpace(strings.TrimPrefix(content, command.Prefix()))
	args := strings.Split(strings.ReplaceAll(arg, "　", " "), " ")
	return strconv.Atoi(args[0])
}

type recruitState struct {
	maxCapacity  int
	author       recruit.UserID
	joinUsers    []recruit.UserID
	declineUsers []recruit.UserID
}

func InitState(authorID string, maxCapacity int) *recruitState {
	return &recruitState{
		maxCapacity:  maxCapacity,
		author:       recruit.UserID(authorID),
		joinUsers:    []recruit.UserID{recruit.UserID(authorID)},
		declineUsers: []recruit.UserID{},
	}
}

func fromRecruitView(view *recruit.RecruitView) *recruitState {
	return &recruitState{
		maxCapacity:  view.Meta.MaxCapacity,
		author:       view.Meta.AuthorID,
		joinUsers:    view.JoinedUsers,
		declineUsers: view.DeclinedUsers,
	}
}

func (state *recruitState) toJoinUsersString() string {
	return state.toUsersString(state.joinUsers)
}

func (state *recruitState) toDeclineUsersString() string {
	return state.toUsersString(state.declineUsers)
}

func (state *recruitState) toUsersString(userIds []recruit.UserID) string {
	var b strings.Builder
	for i, id := range userIds {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(discord.FormatMention(string(id)))
	}
	return b.String()
}

func (state *recruitState) toEmbed() *discordgo.MessageEmbed {
	author := discord.FormatMention(string(state.author))
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📢 募集開始 @%d", state.maxCapacity),
		Description: fmt.Sprintf("%s が募集を始めました", author),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   joinLabel,
				Value:  state.toJoinUsersString(),
				Inline: true,
			},
			{
				Name:   declineLabel,
				Value:  state.toDeclineUsersString(),
				Inline: true,
			},
		},
		Color: 0xffa500,
	}
}

func (state *recruitState) toComponent() discordgo.ActionsRow {
	joinCustomID, _ := encodeCustomID(map[string]string{
		customIDKey: interactionJoin.toString(),
	})
	declineCustomID, _ := encodeCustomID(map[string]string{
		customIDKey: interactionDecline.toString(),
	})

	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    joinLabel,
				Style:    discordgo.PrimaryButton,
				CustomID: joinCustomID,
			},
			discordgo.Button{
				Label:    declineLabel,
				Style:    discordgo.SecondaryButton,
				CustomID: declineCustomID,
			},
		},
	}
}

func (state *recruitState) toMessageContent() *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{state.toEmbed()},
		Components: []discordgo.MessageComponent{state.toComponent()},
	}
}

type baseInteractionCommand struct {
	customID string
}

func (command *baseInteractionCommand) CustomID() string {
	return command.customID
}

func (command *baseInteractionCommand) MatchCustomID(customID string) bool {
	items, err := decodeCustomID(customID)
	if err != nil {
		return false
	}
	return items[customIDKey] == command.customID
}

func encodeCustomID(items map[string]string) (string, error) {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb := strings.Builder{}
	for _, key := range keys {
		value := items[key]
		fmt.Fprintf(&sb, "%d:%s%d:%s", len(key), key, len(value), value)
	}

	result := sb.String()
	if len(result) > 100 {
		return "", fmt.Errorf("custom ID is size over: %s", result)
	}
	return result, nil
}

func decodeCustomID(encodedStr string) (map[string]string, error) {
	data := make(map[string]string)
	i := 0

	for i < len(encodedStr) {
		// キーの長さを読む
		colon := strings.Index(encodedStr[i:], ":")
		if colon == -1 {
			return nil, fmt.Errorf("invalid format: missing colon for key length")
		}
		keyLen, err := strconv.Atoi(encodedStr[i : i+colon])
		if err != nil {
			return nil, fmt.Errorf("invalid key length: %w", err)
		}
		i += colon + 1

		// キーを読む
		if i+keyLen > len(encodedStr) {
			return nil, fmt.Errorf("invalid format: key length exceeds string")
		}
		key := encodedStr[i : i+keyLen]
		i += keyLen

		// 値の長さを読む
		colon = strings.Index(encodedStr[i:], ":")
		if colon == -1 {
			return nil, fmt.Errorf("invalid format: missing colon for value length")
		}
		valLen, err := strconv.Atoi(encodedStr[i : i+colon])
		if err != nil {
			return nil, fmt.Errorf("invalid value length: %w", err)
		}
		i += colon + 1

		// 値を読む
		if i+valLen > len(encodedStr) {
			return nil, fmt.Errorf("invalid format: value length exceeds string")
		}
		value := encodedStr[i : i+valLen]
		i += valLen

		data[key] = value
	}

	return data, nil
}

func (command *baseInteractionCommand) editInteractionResponse(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
	message string,
) {
	command.editInteractionResponseWithComponent(session, interaction, message, nil)
}

func (command *baseInteractionCommand) editInteractionResponseWithComponent(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
	message string,
	component *[]discordgo.MessageComponent,
) {
	_, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Content:    ptr(message),
		Components: component,
	})
	if err != nil {
		log.Printf("failed to edit interaction response: %v", err)
	}
}

type participantActionCommand struct {
	baseInteractionCommand
	service    *recruit.RecruitUsecase
	actionType recruit.ParticipantStatus
}

func NewJoinRecruitCommand(service *recruit.RecruitUsecase) *participantActionCommand {
	return &participantActionCommand{
		service:    service,
		actionType: recruit.ParticipantStatusJoined,
		baseInteractionCommand: baseInteractionCommand{
			customID: interactionJoin.toString(),
		},
	}
}

func NewDeclineRecruitCommand(service *recruit.RecruitUsecase) *participantActionCommand {
	return &participantActionCommand{
		service:    service,
		actionType: recruit.ParticipantStatusDeclined,
		baseInteractionCommand: baseInteractionCommand{
			customID: interactionDecline.toString(),
		},
	}
}

func NewCancelRecruitCommand(service *recruit.RecruitUsecase) *participantActionCommand {
	return &participantActionCommand{
		service:    service,
		actionType: recruit.ParticipantStatusCanceled,
		baseInteractionCommand: baseInteractionCommand{
			customID: interactionCancel.toString(),
		},
	}
}

func (command *participantActionCommand) InteractionType() discordgo.InteractionType {
	return discordgo.InteractionMessageComponent
}

func (command *participantActionCommand) Handle(session *discordgo.Session, interaction *discordgo.Interaction) error {
	// 3秒以内に応答する必要があるのでBOT待機メッセージで返答
	// キャンセルボタンの場合は元のエフェメラルメッセージを更新対象にする
	var responseType discordgo.InteractionResponseType
	if command.actionType == recruit.ParticipantStatusCanceled {
		responseType = discordgo.InteractionResponseDeferredMessageUpdate
	} else {
		responseType = discordgo.InteractionResponseDeferredChannelMessageWithSource
	}

	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	// Note: interaction.UserはDMでのインタラクションユーザーが入る
	// サーバーでのインタラクションユーザーはinteraction.Member.User
	actorID := recruit.UserID(interaction.Member.User.ID)
	channelID := recruit.ChannelID(interaction.ChannelID)
	messageID, err := command.extractMessageID(interaction)
	if err != nil {
		return err
	}

	ctx, cancel := createContextWithTimeout()
	defer cancel()

	// ビジネスロジック呼び出し
	result, err := command.executeAction(ctx, channelID, messageID, actorID)
	if err != nil {
		return command.handleActionError(session, interaction, err)
	}

	// 募集メッセージの編集
	if err := command.updateRecruitMessage(session, result.CurrentView); err != nil {
		return err
	}

	// BOT待機メッセージを削除
	_ = session.InteractionResponseDelete(interaction)
	// 追加メッセージの送信またはエフェメラルメッセージの完了処理
	return command.sendFollowUpMessage(session, result, actorID)
}

func (command *participantActionCommand) extractMessageID(interaction *discordgo.Interaction) (recruit.MessageID, error) {
	switch command.actionType {
	case recruit.ParticipantStatusCanceled:
		customID := interaction.MessageComponentData().CustomID
		items, err := decodeCustomID(customID)
		if err != nil {
			return "", err
		}
		return recruit.MessageID(items[messageIDKey]), nil
	default:
		return recruit.MessageID(interaction.Message.ID), nil
	}
}

func (command *participantActionCommand) executeAction(
	ctx context.Context,
	channelID recruit.ChannelID,
	messageID recruit.MessageID,
	userID recruit.UserID,
) (*recruit.ParticipantStatusChangeResult, error) {
	switch command.actionType {
	case recruit.ParticipantStatusJoined:
		return command.service.Join(ctx, channelID, messageID, userID)
	case recruit.ParticipantStatusDeclined:
		return command.service.Decline(ctx, channelID, messageID, userID)
	case recruit.ParticipantStatusCanceled:
		return command.service.Cancel(ctx, channelID, messageID, userID)
	default:
		return nil, fmt.Errorf("invalid action type: %s", command.actionType)
	}
}

func (command *participantActionCommand) handleActionError(session *discordgo.Session, interaction *discordgo.Interaction, err error) error {
	if errors.Is(err, recruit.ErrAuthorCannotJoin) {
		command.sendAuthorControlPanel(session, interaction)
		return nil
	}
	if errors.Is(err, recruit.ErrAlreadyJoined) || errors.Is(err, recruit.ErrAlreadyDeclined) {
		command.sendParticipantControlPanel(session, interaction)
		return nil
	}
	command.editInteractionResponse(session, interaction, "❗処理中に問題が発生しました。")
	return err
}

func (command *participantActionCommand) sendAuthorControlPanel(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
) {
	message := "作成者は参加/辞退できません。\n募集を削除する場合はボタンを押下してください。"
	deleteCustomID, _ := encodeCustomID(map[string]string{
		customIDKey:  interactionClose.toString(),
		messageIDKey: interaction.Message.ID,
	})
	button := &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🗑️ 削除",
					Style:    discordgo.DangerButton,
					CustomID: deleteCustomID,
				},
			},
		},
	}
	command.editInteractionResponseWithComponent(session, interaction, message, button)
}

func (command *participantActionCommand) sendParticipantControlPanel(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
) {
	message := "既に参加済み/辞退済みです。\nキャンセルする場合はボタンを押下してください。"
	cancelCustomID, _ := encodeCustomID(map[string]string{
		customIDKey:  interactionCancel.toString(),
		messageIDKey: interaction.Message.ID,
	})
	button := &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "❌ キャンセル",
					Style:    discordgo.DangerButton,
					CustomID: cancelCustomID,
				},
			},
		},
	}
	command.editInteractionResponseWithComponent(session, interaction, message, button)
}

func (command *participantActionCommand) updateRecruitMessage(session *discordgo.Session, view *recruit.RecruitView) error {
	state := fromRecruitView(view)
	_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    string(view.Meta.ChannelID),
		ID:         string(view.Meta.MessageID),
		Embeds:     &[]*discordgo.MessageEmbed{state.toEmbed()},
		Components: &[]discordgo.MessageComponent{state.toComponent()},
	})
	return err
}

func (command *participantActionCommand) sendFollowUpMessage(
	session *discordgo.Session,
	result *recruit.ParticipantStatusChangeResult,
	actorID recruit.UserID,
) error {
	view := result.CurrentView
	switch command.actionType {
	case recruit.ParticipantStatusJoined:
		// 参加メッセージの送信
		return command.replyRecruitMessage(session, view, createJoinMessage(actorID, view))
	case recruit.ParticipantStatusDeclined, recruit.ParticipantStatusCanceled:
		// 参加済みから辞退/キャンセルに変更された場合のみ通知
		if result.PreviousStatus != nil && *result.PreviousStatus == recruit.ParticipantStatusJoined {
			message := fmt.Sprintf(
				"%s が参加を取り消しました。 @%d",
				discord.FormatMention(string(actorID)),
				view.RemainingSlots(),
			)
			return command.replyRecruitMessage(session, view, message)
		}
		return nil
	default:
		return nil
	}
}

func (command *participantActionCommand) replyRecruitMessage(
	session *discordgo.Session,
	view *recruit.RecruitView,
	content string,
) error {
	_, err := session.ChannelMessageSendComplex(
		string(view.Meta.ChannelID),
		&discordgo.MessageSend{
			Content: content,
			Reference: &discordgo.MessageReference{
				MessageID: string(view.Meta.MessageID),
			},
		},
	)
	return err
}

func createJoinMessage(actorID recruit.UserID, view *recruit.RecruitView) string {
	baseContent := fmt.Sprintf(
		"%s が参加しました。",
		discord.FormatMention(string(actorID)),
	)

	if !view.IsFull() {
		return fmt.Sprintf(
			"%s @%d",
			baseContent,
			view.RemainingSlots(),
		)
	}

	if view.ExtraCount() == 0 {
		var userIds []string
		for _, u := range view.JoinedUsers {
			userIds = append(userIds, discord.FormatMention(string(u)))
		}

		return fmt.Sprintf(
			"%s\n\n**[募集終了]**\n%s",
			baseContent,
			strings.Join(userIds, " "),
		)
	}

	return baseContent
}

func ptr(s string) *string {
	return &s
}

type closeRecruitCommand struct {
	baseInteractionCommand
	service *recruit.RecruitUsecase
}

func NewCloseRecruitCommand(service *recruit.RecruitUsecase) *closeRecruitCommand {
	return &closeRecruitCommand{
		service: service,
		baseInteractionCommand: baseInteractionCommand{
			customID: interactionClose.toString(),
		},
	}
}

func (command *closeRecruitCommand) InteractionType() discordgo.InteractionType {
	return discordgo.InteractionMessageComponent
}

func (command *closeRecruitCommand) Handle(session *discordgo.Session, interaction *discordgo.Interaction) error {
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	if err != nil {
		return err
	}

	// 削除ボタンに埋め込まれた値をデコード
	customID := interaction.MessageComponentData().CustomID
	items, err := decodeCustomID(customID)
	if err != nil {
		return err
	}

	actorID := recruit.UserID(interaction.Member.User.ID)
	channelID := recruit.ChannelID(interaction.ChannelID)
	recruitMessageIDStr := items[messageIDKey]
	recruitMessageID := recruit.MessageID(recruitMessageIDStr)

	ctx, cancel := createContextWithTimeout()
	defer cancel()

	// 削除ロジック実行
	err = command.service.Close(ctx, channelID, recruitMessageID, actorID)
	if err != nil {
		return err
	}

	// 元の募集メッセージの内容を削除用に差し替え
	_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    interaction.ChannelID,
		ID:         recruitMessageIDStr,
		Content:    ptr("募集は削除されました。"),
		Embeds:     &[]*discordgo.MessageEmbed{},
		Components: &[]discordgo.MessageComponent{},
	})
	// 削除ボタンインタラクションを削除
	_ = session.InteractionResponseDelete(interaction)

	return err
}

func createContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
