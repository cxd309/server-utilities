# Marks MP4 files as podcasts. iTunes silently refuses a file whose media kind is
# Podcast unless it also has this flag, and ffmpeg can't write it, so mutagen does.
import sys

from mutagen.mp4 import MP4

for path in sys.argv[1:]:
    tags = MP4(path)
    tags["pcst"] = True
    tags.save()
