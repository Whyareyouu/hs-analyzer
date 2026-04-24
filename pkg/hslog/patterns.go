package hslog

import "regexp"

var (
	LoadingScreenSceneLoadedPattern = regexp.MustCompile(`LoadingScreen\.OnSceneLoaded\(\) - prevMode=(\S+) currMode=(\S+)`)
	LoadingScreenBoardPattern       = regexp.MustCompile(`LoadingScreen\.DisableTransitionUnfriendlyStuff\(\) - (\w+)\(Clone\)`)
	PowerGameTypePattern            = regexp.MustCompile(`GameType=(GT_[A-Z_]+)`)
	PowerTagChangePattern           = regexp.MustCompile(`TAG_CHANGE Entity=(.+?) tag=([A-Z0-9_]+) value=([^\s]+)`)
	PowerFullEntityPattern          = regexp.MustCompile(`FULL_ENTITY - (?:Creating ID=(\d+)|Updating \[.* id=(\d+) .*] CardID=([A-Z0-9_]+))`)
	PowerShowEntityPattern          = regexp.MustCompile(`SHOW_ENTITY - Updating Entity=.* CardID=([A-Z0-9_]+)`)
	PowerBlockStartPattern          = regexp.MustCompile(`BLOCK_START BlockType=([A-Z_]+)`)
	PowerEntityPattern              = regexp.MustCompile(`Entity=\[entityName=(.+?) id=(\d+) zone=([A-Z_]+) zonePos=(\d+) cardId=([^ ]*) player=(\d+)\]`)
	PowerTargetPattern              = regexp.MustCompile(`Target=\[entityName=(.+?) id=(\d+) zone=([A-Z_]+) zonePos=(\d+) cardId=([^ ]*) player=(\d+)\]`)
	PowerTriggerKeywordPattern      = regexp.MustCompile(`TriggerKeyword=([A-Z0-9_]+)`)
)
