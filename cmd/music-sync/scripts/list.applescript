-- Prints one line per file track in the library: "persistentID<tab>path".
-- The path is empty when the Music app can't find the file.
on run argv
	tell application "Music"
		if (count of file tracks of library playlist 1) is 0 then return ""
		set ids to persistent ID of every file track of library playlist 1
		set locs to location of every file track of library playlist 1
	end tell

	set output to {}
	repeat with i from 1 to count of ids
		set loc to item i of locs
		if loc is missing value then
			set p to ""
		else
			set p to POSIX path of loc
		end if
		set end of output to (item i of ids) & tab & p
	end repeat

	set AppleScript's text item delimiters to linefeed
	return output as text
end run