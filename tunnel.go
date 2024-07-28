package lokal

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fatih/color"
)

type TunnelType string

const (
	TunnelTypeHTTP TunnelType = "HTTP"
)

type Tunnel struct {
	Lokal *Lokal `json:",omitempty"`

	ID                string     `json:",omitempty"`
	Name              string     `json:"name"`
	TunnelType        TunnelType `json:"tunnel_type"`
	LocalAddress      string     `json:"local_address"`
	ServerID          string     `json:"server_id"`
	AddressTunnel     string     `json:"address_tunnel"`
	AddressTunnelPort int64      `json:"address_tunnel_port"`
	AddressPublic     string     `json:"address_public"`
	AddressMdns       string     `json:"address_mdns"`
	Inspect           bool       `json:"inspect"`
	Options           Options    `json:"options"`

	ignoreDuplicate bool
	startupBanner   bool
}

type Options struct {
	BasicAuth            []string `json:"basic_auth"`
	CIDRAllow            []string `json:"cidr_allow"`
	CIDRDeny             []string `json:"cidr_deny"`
	RequestHeaderAdd     []string `json:"request_header_add"`
	RequestHeaderRemove  []string `json:"request_header_remove"`
	ResponseHeaderAdd    []string `json:"response_header_add"`
	ResponseHeaderRemove []string `json:"response_header_remove"`
	HeaderKey            []string `json:"header_key"`
}

func (l *Lokal) NewTunnel() *Tunnel {
	return &Tunnel{Lokal: l}
}

func (t *Tunnel) SetLocalAddress(localAddress string) *Tunnel {
	t.LocalAddress = localAddress
	return t
}

func (t *Tunnel) SetTunnelType(tunnelType TunnelType) *Tunnel {
	t.TunnelType = tunnelType
	return t
}

func (t *Tunnel) SetInspection(inspect bool) *Tunnel {
	t.Inspect = inspect
	return t
}

func (t *Tunnel) SetLANAddress(lanAddress string) *Tunnel {
	lanAddress = strings.TrimSuffix(lanAddress, ".local")
	t.AddressMdns = lanAddress
	return t
}

func (t *Tunnel) SetPublicAddress(publicAddress string) *Tunnel {
	t.AddressPublic = publicAddress
	return t
}

func (t *Tunnel) SetName(name string) *Tunnel {
	t.Name = name
	return t
}

func (t *Tunnel) IgnoreDuplicate() *Tunnel {
	t.ignoreDuplicate = true
	return t
}

func (t *Tunnel) ShowStartupBanner() *Tunnel {
	t.startupBanner = true
	return t
}

func (t *Tunnel) Create() (*Tunnel, error) {
	if t.AddressMdns == "" && t.AddressPublic == "" {
		return nil, errors.New("please enable either lan address or random/custom public url")
	}

	resp := struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Tunnel  []Tunnel `json:"data"`
	}{}

	_, err := t.Lokal.rest.
		R().
		SetBody(t).
		SetResult(&resp).
		SetError(&resp).
		Post("/api/tunnel/start")
	if err != nil {
		return nil, err
	}

	if len(resp.Tunnel) == 0 {
		return nil, errors.New("tunnel creation failing")
	}

	if !resp.Success {
		if t.ignoreDuplicate && strings.HasSuffix(resp.Message, "address is already being used") {
			t.AddressPublic = resp.Tunnel[0].AddressPublic
			t.AddressMdns = resp.Tunnel[0].AddressMdns
			t.ID = resp.Tunnel[0].ID

			t.showStartupBanner()
			return t, nil
		}
		return nil, errors.New(resp.Message)
	}

	t.AddressPublic = resp.Tunnel[0].AddressPublic
	t.AddressMdns = resp.Tunnel[0].AddressMdns
	t.ID = resp.Tunnel[0].ID

	t.showStartupBanner()

	return t, nil
}

func (t *Tunnel) GetLANAddress() (string, error) {
	if t.AddressMdns == "" {
		return "", errors.New("lan address is not being set")
	}

	if !strings.HasSuffix(t.AddressMdns, ".local") {
		return t.AddressMdns + ".local", nil
	}

	return t.AddressMdns, nil
}

func (t *Tunnel) GetPublicAddress() (string, error) {
	if t.AddressPublic == "" {
		return "", errors.New("public address is not requested by client")
	}

	if t.AddressPublic == "" {
		return "", errors.New("unable to assign public address")
	}

	if t.TunnelType != "HTTP" {
		if !strings.Contains(t.AddressPublic, ":") {
			go t.updatePublicURLPort()
			return "", errors.New("tunnel is using a random port, but it has not been assigned yet. please try again later")
		}
	}

	return t.AddressPublic, nil
}

func (t *Tunnel) updatePublicURLPort() error {

	resp := struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Tunnel  []Tunnel `json:"data"`
	}{}

	_, err := t.Lokal.rest.R().
		SetResult(&resp).
		SetError(&resp).
		Get("/api/tunnel/info/" + t.ID)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	if len(resp.Tunnel) == 0 {
		return errors.New("could not get tunnel info")
	}

	if !strings.Contains(resp.Tunnel[0].AddressPublic, ":") {
		return errors.New("could not get assigned port")
	}

	t.AddressPublic = resp.Tunnel[0].AddressPublic

	return nil
}

func (t *Tunnel) showStartupBanner() {
	if !t.startupBanner {
		return
	}

	var banner = `
    __       _         _             
   / /  ___ | | ____ _| |  ___  ___  
  / /  / _ \| |/ / _  | | / __|/ _ \ 
 / /__| (_) |   < (_| | |_\__ \ (_) |
 \____/\___/|_|\_\__,_|_(_)___/\___/ `

	colors := []func(format string, a ...interface{}) string{
		color.HiMagentaString,
		color.HiBlueString,
		color.HiCyanString,
		color.HiGreenString,
		color.HiRedString,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println(colors[r.Intn(len(colors))](banner))
	fmt.Println()
	fmt.Println(color.RedString("Minimum Lokal Client"), "\t"+ServerMinVersion)
	if val, err := t.GetPublicAddress(); err == nil {
		fmt.Println(color.CyanString("Public Address"), "\t\thttps://"+val)
	}
	if val, err := t.GetLANAddress(); err == nil {
		fmt.Println(color.GreenString("LAN Address"), "\t\thttps://"+val)
	}
	fmt.Println()
}
