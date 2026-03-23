package domain

type Target string

const (
	TargetAll      Target = "all"
	TargetMihomo   Target = "mihomo"
	TargetStash    Target = "stash"
	TargetLoon     Target = "loon"
	TargetSurge    Target = "surge"
	TargetEasytier Target = "easytier"
)

type Mode string

const (
	ModeStandard Mode = "standard"
	ModeFull     Mode = "full"
	ModeTun      Mode = "tun"
	ModeOverride Mode = "override"
	ModeArgs     Mode = "args"
)
