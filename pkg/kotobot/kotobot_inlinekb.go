package kotobot

type TInlineKeyboard struct {
	inlineKbMarkup TInlineKeyboardMarkup
}

func NewInlineKB(rows uint8, cols uint8) TInlineKeyboard {
	var inlineKb TInlineKeyboard
	if rows > 0 {
		inlineKb.inlineKbMarkup.InlineKeyboard = make([][]TInlineKeyboardButton, rows)
		if cols > 0 {
			for r := range inlineKb.inlineKbMarkup.InlineKeyboard {
				inlineKb.inlineKbMarkup.InlineKeyboard[r] = make([]TInlineKeyboardButton, cols)
			}
		}
	}
	return inlineKb
}

func (ILKB *TInlineKeyboard) SetButton(row uint8, col uint8, text string, cbdata *string) {
	ILKB.inlineKbMarkup.InlineKeyboard[row][col] = TInlineKeyboardButton{Text: text, CallbackData: cbdata}
}

func (ILKB TInlineKeyboard) Egress() (*TInlineKeyboardMarkup) {
	return &ILKB.inlineKbMarkup
}
