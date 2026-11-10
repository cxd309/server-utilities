-- Adds each file path given as an argument to the library.
on run argv
	set toAdd to {}
	repeat with p in argv
		set end of toAdd to (POSIX file (contents of p))
	end repeat

	tell application "Music"
		with timeout of 600 seconds
			add toAdd to library playlist 1
		end timeout
	end tell
end run