-- Removes each persistent ID given as an argument from the library.
-- The files on disk are not touched.
on run argv
	tell application "Music"
		repeat with pid in argv
			repeat with t in (file tracks of library playlist 1 whose persistent ID is (contents of pid))
				delete t
			end repeat
		end repeat
	end tell
end run