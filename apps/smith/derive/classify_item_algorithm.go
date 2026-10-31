package derive

import (
	"errors"

	"github.com/monosmith-app/monosmith/apps/smith/derivefile"
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Classify compares what a renderer wants (f), what is on disk (disk, exists)
// and what the record says (entry, nil when unlisted), and says what derive
// finds. It reads and writes nothing.
func Classify(f protocol.DerivedFile, entry *derivelock.Entry, disk []byte, exists bool) Item {
	item := Item{File: f, Owner: derivelock.OwnerSmith, Exists: exists, Entry: entry}
	if entry != nil {
		item.Owner = entry.Owner
	}
	wanted, err := derivefile.Wanted(f)
	if err != nil {
		return broken(item, err.Error())
	}
	item.Wanted = wanted
	if entry != nil && entry.Owner == derivelock.OwnerYou {
		item.Status = StatusEjected
		item.Detail = "owned by you: smith does not touch it"
		return item
	}
	if !exists {
		return classifyMissing(item, entry)
	}
	present, found, err := derivefile.Present(f, disk)
	if err != nil {
		return broken(item, err.Error())
	}
	if !found {
		return broken(item, "the file has no smith:begin and smith:end markers for "+f.Renderer+": add them around the lines smith owns")
	}
	item.Present = present
	if gone, err := derivefile.Absent(f, disk); err == nil && gone {
		return classifyDeleted(item, disk)
	}
	return classifyPresent(item, entry, disk)
}

func classifyMissing(item Item, entry *derivelock.Entry) Item {
	merged, err := derivefile.Merge(item.File, nil, false)
	if errors.Is(err, derivefile.ErrNoMarkers) {
		return broken(item, "the file is gone and a region cannot be created without it")
	}
	if err != nil {
		return broken(item, err.Error())
	}
	item.Merged = merged
	item.Status = StatusMissing
	item.Detail = "the file is missing; derive writes it"
	if entry == nil {
		item.Detail = "the renderer derives it and it does not exist yet; derive writes it"
	}
	return item
}

// classifyDeleted handles a file whose owned part was deleted (an empty region,
// none of the owned keys): derive puts it back, as for a missing file.
func classifyDeleted(item Item, disk []byte) Item {
	merged, err := derivefile.Merge(item.File, disk, true)
	if err != nil {
		return broken(item, err.Error())
	}
	item.Merged = merged
	item.Status = StatusMissing
	item.Detail = "what the renderer owns was deleted from the file; derive writes it back"
	return item
}

func classifyPresent(item Item, entry *derivelock.Entry, disk []byte) Item {
	wantedHash, presentHash := derivelock.Hash(item.Wanted), derivelock.Hash(item.Present)
	switch {
	case entry == nil && presentHash == wantedHash:
		item.Status = StatusUnrecorded
		item.Detail = "the file is as the renderer wants it; derive records it"
	case entry == nil:
		item.Status = StatusDiverged
		item.Detail = "the file exists and differs from what the renderer derives; derive leaves it alone"
	case presentHash == wantedHash:
		item.Status = StatusCurrent
		item.Detail = "up to date"
		if entry.Hash != wantedHash || entry.Wanted != wantedHash || entry.Version != item.File.Version || entry.Renderer != item.File.Renderer || entry.Mode != item.File.Mode {
			item.Status = StatusUnrecorded
			item.Detail = "the file is as the renderer wants it; derive refreshes the record"
		}
	case presentHash != entry.Hash:
		item.Status = StatusDrift
		item.Detail = "edited by hand since smith wrote it; derive does not overwrite it"
	case entry.Hash == entry.Wanted:
		item.Status = StatusStale
		item.Detail = "the renderer's inputs changed; derive rewrites it"
	case wantedHash == entry.Wanted:
		item.Status = StatusCurrent
		item.Detail = "a hand edit that was adopted stands"
	default:
		item.Status = StatusDrift
		item.Detail = "an adopted hand edit conflicts with changed inputs; derive does not overwrite it"
	}
	if item.Status == StatusStale || item.Status == StatusDrift || item.Status == StatusDiverged {
		merged, err := derivefile.Merge(item.File, disk, true)
		if err != nil {
			return broken(item, err.Error())
		}
		item.Merged = merged
	}
	return item
}

func broken(item Item, detail string) Item {
	item.Status = StatusBroken
	item.Detail = detail
	return item
}
