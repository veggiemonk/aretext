package input

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/text"
)

// Action is a function that mutates the editor state.
type Action func(*state.EditorState)

// EmptyAction is an action that does nothing.
func EmptyAction(s *state.EditorState) {}

func countArgOrDefault(countArg *uint64, defaultCount uint64) uint64 {
	if countArg != nil {
		return *countArg
	} else {
		return defaultCount
	}
}

func CursorLeft(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorBack(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevChar(params.TextTree, 1, params.CursorPos)
	})
}

func CursorRight(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorRightIncludeEndOfLineOrFile(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
	})
}

func CursorUp(s *state.EditorState) {
	state.MoveCursorToLineAbove(s, 1)
}

func CursorDown(s *state.EditorState) {
	state.MoveCursorToLineBelow(s, 1)
}

func CursorNextWordStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordStart(params.TextTree, params.TokenTree, params.CursorPos, false)
	})
}

func CursorPrevWordStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func CursorNextWordEnd(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func CursorPrevParagraph(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevParagraph(params.TextTree, params.CursorPos)
	})
}

func CursorNextParagraph(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextParagraph(params.TextTree, params.CursorPos)
	})
}

func ScrollUp(config Config) Action {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	return func(s *state.EditorState) {
		// Move the cursor to the start of a line above, then scroll up.
		// (We don't scroll the view, because that happens automatically after every action.)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.StartOfLineAbove(params.TextTree, scrollLines, params.CursorPos)
		})
		state.ScrollViewByNumLines(s, text.ReadDirectionBackward, scrollLines)
	}
}

func ScrollDown(config Config) Action {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	return func(s *state.EditorState) {
		// Move the cursor to the start of a line below, then scroll down.
		// (We don't scroll the view, because that happens automatically after every action.)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, scrollLines, params.CursorPos)
		})
		state.ScrollViewByNumLines(s, text.ReadDirectionForward, scrollLines)
	}
}

func CursorLineStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
	})
}

func CursorLineStartNonWhitespace(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func CursorLineEnd(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, false, params.CursorPos)
	})
}

func CursorLineEndIncludeEndOfLineOrFile(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
	})
}

func CursorStartOfLineNum(countArg *uint64) Action {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if countArg != nil && *countArg > 0 {
		lineNum = *countArg - 1
	}

	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.StartOfLineNum(params.TextTree, lineNum)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func CursorStartOfLastLine(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		lineStartPos := locate.StartOfLastLine(params.TextTree)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func EnterInsertMode(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
}

func EnterInsertModeAtStartOfLine(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorLineStartNonWhitespace(s)
}

func EnterInsertModeAtNextPos(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorRightIncludeEndOfLineOrFile(s)
}

func EnterInsertModeAtEndOfLine(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorLineEndIncludeEndOfLineOrFile(s)
}

func ReturnToNormalMode(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeNormal)
}

func ReturnToNormalModeAfterInsert(s *state.EditorState) {
	state.ClearAutoIndentWhitespaceLine(s, func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAtPos(params.TextTree, params.CursorPos)
	})
	CursorLeft(s)
	state.SetInputMode(s, state.InputModeNormal)
}

func InsertRune(r rune) Action {
	return func(s *state.EditorState) {
		state.InsertRune(s, r)
	}
}

func InsertNewlineAndUpdateAutoIndentWhitespace(s *state.EditorState) {
	state.InsertNewline(s)
	state.ClearAutoIndentWhitespaceLine(s, func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
	})
}

func InsertTab(s *state.EditorState) {
	state.InsertTab(s)
}

func DeletePrevChar(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		prevInLinePos := locate.PrevCharInLine(params.TextTree, 1, true, params.CursorPos)
		prevAutoIndentPos := locate.PrevAutoIndent(
			params.TextTree,
			params.AutoIndentEnabled,
			params.TabSize,
			params.CursorPos)
		if prevInLinePos < prevAutoIndentPos {
			return prevInLinePos
		} else {
			return prevAutoIndentPos
		}
	})
}

func BeginNewLineBelow(s *state.EditorState) {
	CursorLineEndIncludeEndOfLineOrFile(s)
	state.InsertNewline(s)
	state.SetInputMode(s, state.InputModeInsert)
}

func BeginNewLineAbove(s *state.EditorState) {
	state.BeginNewLineAbove(s)
	EnterInsertMode(s)
}

func JoinLines(s *state.EditorState) {
	state.JoinLines(s)
}

func DeleteLines(countArg *uint64) Action {
	count := countArgOrDefault(countArg, 0)
	if count > 0 {
		count--
	}
	return func(s *state.EditorState) {
		targetLoc := func(params state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, count, params.CursorPos)
		}
		state.DeleteLines(s, targetLoc, false, false)
		CursorLineStartNonWhitespace(s)
	}
}

func DeletePrevCharInLine(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func DeleteNextCharInLine(countArg *uint64) Action {
	count := countArgOrDefault(countArg, 1)
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, count, true, params.CursorPos)
		})
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteDown(s *state.EditorState) {
	targetLineLoc := func(params state.LocatorParams) uint64 {
		return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
	}
	state.DeleteLines(s, targetLineLoc, true, false)
	CursorLineStartNonWhitespace(s)
}

func DeleteUp(s *state.EditorState) {
	targetLineLoc := func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
	}
	state.DeleteLines(s, targetLineLoc, true, false)
	CursorLineStartNonWhitespace(s)
}

func DeleteToEndOfLine(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
	})
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
	})
}

func DeleteToStartOfLine(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
	})
}

func DeleteToStartOfLineNonWhitespace(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func DeleteToStartOfNextWord(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordStartInLine(params.TextTree, params.TokenTree, params.CursorPos)
	})
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
	})
}

func DeleteAWord(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	})
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.CurrentWordEndWithTrailingWhitespace(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func DeleteInnerWord(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	})
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.CurrentWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func ChangeToStartOfNextWord(s *state.EditorState) {
	state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordStartInLine(params.TextTree, params.TokenTree, params.CursorPos)
	})
	EnterInsertMode(s)
}

func ChangeAWord(s *state.EditorState) {
	DeleteAWord(s)
	EnterInsertMode(s)
}

func ChangeInnerWord(s *state.EditorState) {
	DeleteInnerWord(s)
	EnterInsertMode(s)
}

func ReplaceCharacter(inputEvents []*tcell.EventKey) Action {
	if len(inputEvents) == 0 {
		// This should never happen if the parser rule is configured correctly.
		panic("Replace chars expects at least one input event")
	}

	lastInput := inputEvents[len(inputEvents)-1]
	var newRune rune
	if lastInput.Key() == tcell.KeyEnter {
		newRune = '\n'
	} else if lastInput.Key() == tcell.KeyRune {
		newRune = lastInput.Rune()
	} else {
		log.Printf("Unsupported input for replace character command\n")
		return EmptyAction
	}

	newText := string([]rune{newRune})
	return func(s *state.EditorState) {
		state.ReplaceChar(s, newText)
	}
}

func ToggleCaseAtCursor(s *state.EditorState) {
	state.ToggleCaseAtCursor(s)
}

func IndentLine(s *state.EditorState) {
	state.IndentLineAtCursor(s)
}

func OutdentLine(s *state.EditorState) {
	state.OutdentLineAtCursor(s)
}

func CopyToStartOfNextWord(s *state.EditorState) {
	startLoc := func(params state.LocatorParams) uint64 {
		return params.CursorPos
	}
	endLoc := func(params state.LocatorParams) uint64 {
		return locate.NextWordStartInLine(params.TextTree, params.TokenTree, params.CursorPos)
	}
	state.CopyRegion(s, startLoc, endLoc)
}

func CopyAWord(s *state.EditorState) {
	startLoc := func(params state.LocatorParams) uint64 {
		return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	}
	endLoc := func(params state.LocatorParams) uint64 {
		return locate.CurrentWordEndWithTrailingWhitespace(params.TextTree, params.TokenTree, params.CursorPos)
	}
	state.CopyRegion(s, startLoc, endLoc)
}

func CopyInnerWord(s *state.EditorState) {
	startLoc := func(params state.LocatorParams) uint64 {
		return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	}
	endLoc := func(params state.LocatorParams) uint64 {
		return locate.CurrentWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
	}
	state.CopyRegion(s, startLoc, endLoc)
}

func CopyLines(s *state.EditorState) {
	state.CopyLine(s)
}

func PasteAfterCursor(s *state.EditorState) {
	state.PasteAfterCursor(s, clipboard.PageDefault)
}

func PasteBeforeCursor(s *state.EditorState) {
	state.PasteBeforeCursor(s, clipboard.PageDefault)
}

func ShowCommandMenu(config Config) Action {
	return func(s *state.EditorState) {
		// This sets the input mode to menu.
		state.ShowMenu(s, state.MenuStyleCommand, commandMenuItems(config))
	}
}

func HideMenuAndReturnToNormalMode(s *state.EditorState) {
	state.HideMenu(s)
}

func ExecuteSelectedMenuItem(s *state.EditorState) {
	// Hides the menu, then executes the menu item action.
	// This usually returns to normal mode, unless the menu item action sets a different mode.
	state.ExecuteSelectedMenuItem(s)
}

func MenuSelectionUp(s *state.EditorState) {
	state.MoveMenuSelection(s, -1)
}

func MenuSelectionDown(s *state.EditorState) {
	state.MoveMenuSelection(s, 1)
}

func AppendRuneToMenuSearch(r rune) Action {
	return func(s *state.EditorState) {
		state.AppendRuneToMenuSearch(s, r)
	}
}

func DeleteRuneFromMenuSearch(s *state.EditorState) {
	state.DeleteRuneFromMenuSearch(s)
}

func StartSearchForward(s *state.EditorState) {
	// This sets the input mode to search.
	state.StartSearch(s, text.ReadDirectionForward)
}

func StartSearchBackward(s *state.EditorState) {
	// This sets the input mode to search.
	state.StartSearch(s, text.ReadDirectionBackward)
}

func AbortSearchAndReturnToNormalMode(s *state.EditorState) {
	state.CompleteSearch(s, false)
}

func CommitSearchAndReturnToNormalMode(s *state.EditorState) {
	state.CompleteSearch(s, true)
}

func AppendRuneToSearchQuery(r rune) Action {
	return func(s *state.EditorState) {
		state.AppendRuneToSearchQuery(s, r)
	}
}

func DeleteRuneFromSearchQuery(s *state.EditorState) {
	state.DeleteRuneFromSearchQuery(s)
}

func FindNextMatch(s *state.EditorState) {
	state.FindNextMatch(s, false)
}

func FindPrevMatch(s *state.EditorState) {
	reverse := true
	state.FindNextMatch(s, reverse)
}

func Undo(s *state.EditorState) {
	state.Undo(s)
}

func Redo(s *state.EditorState) {
	state.Redo(s)
}

func ToggleVisualModeCharwise(s *state.EditorState) {
	state.ToggleVisualMode(s, selection.ModeChar)
}

func ToggleVisualModeLinewise(s *state.EditorState) {
	state.ToggleVisualMode(s, selection.ModeLine)
}

func DeleteSelectionAndReturnToNormalMode(s *state.EditorState) {
	state.DeleteSelection(s, false)
	ReturnToNormalMode(s)
}

func ToggleCaseInSelectionAndReturnToNormalMode(s *state.EditorState) {
	state.ToggleCaseInSelection(s)
	ReturnToNormalMode(s)
}

func IndentSelectionAndReturnToNormalMode(s *state.EditorState) {
	state.IndentSelection(s)
	ReturnToNormalMode(s)
}

func OutdentSelectionAndReturnToNormalMode(s *state.EditorState) {
	state.OutdentSelection(s)
	ReturnToNormalMode(s)
}

func ChangeSelection(s *state.EditorState) {
	state.DeleteSelection(s, true)
	EnterInsertMode(s)
}

func CopySelectionAndReturnToNormalMode(s *state.EditorState) {
	state.CopySelection(s)
	ReturnToNormalMode(s)
}
