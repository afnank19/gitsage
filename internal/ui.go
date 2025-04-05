package internal

import "fmt"

// Moving UI away from the view function

func buildCommitHistoryView(list list) string {
	view := "(3) Branch Commit History"
	view = title.Render(view)
	view += "\n"

	end := min(list.offset+list.height, len(list.items))

	for i := list.offset; i < end; i++ {
		cursor := " "
		currItem := list.items[i]

		if i == list.cursor {
			cursor = pointer.Render(">")
			currItem = cursorStyle.Render(currItem)
		}

		view += fmt.Sprintf("%s %s\n", cursor, currItem)
	}

	return view
}
