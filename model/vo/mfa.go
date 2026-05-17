package vo

import "github.com/evo-lee/evo-sonic/consts"

type MFAFactorAuth struct {
	QRImage    string         `json:"qrImage"`
	OptAuthURL string         `json:"optAuthUrl"`
	MFAKey     string         `json:"mfaKey"`
	MFAType    consts.MFAType `json:"mfaType"`
}
