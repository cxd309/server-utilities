-- Re-reads tags from disk for each persistent ID given as an argument.
on run argv
	tell application "Music"
		repeat with pid in argv
			repeat with t in (file tracks of library playlist 1 whose persistent ID is (contents of pid))
				refresh t
			end repeat
		end repeat
	end tell
end run