package provider

import (
	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/golang-docker-client"
	"github.com/nanobox-io/nanobox/models"
	"github.com/nanobox-io/nanobox/provider"
	"github.com/nanobox-io/nanobox/util/dhcp"
	"github.com/nanobox-io/nanobox/util/display"
	"github.com/nanobox-io/nanobox/util/locker"
)

// Setup ...
type Setup struct {
}

//
func (setup Setup) Run() error {
	display.StartTask("preparing provider")

	locker.GlobalLock()
	defer locker.GlobalUnlock()

	if err := provider.Create(); err != nil {
		display.ErrorTask()
		return err
	}

	display.StopTask()
	
	display.StartTask("booting provider")

	if err := provider.Start(); err != nil {
		display.ErrorTask()
		return err
	}

	if err := setup.saveProvider(); err != nil {
		display.ErrorTask()
		return err
	}

	if err := setup.SetDefaultIP(); err != nil {
		display.ErrorTask()
		return err
	}

	if err := provider.DockerEnv(); err != nil {
		display.ErrorTask()
		return err
	}

	if err := docker.Initialize("env"); err != nil {
		display.ErrorTask()
		return err
	}

	display.StopTask()
	return nil
}

func (setup Setup) saveProvider() error {
	mProvider, _ := models.LoadProvider()

	// if it has already been saved the exit early
	if mProvider.HostIP != "" {
		return nil
	}

	// get a new ip we will use for mounting
	ip, err := dhcp.ReserveGlobal()
	if err != nil {
		lumber.Error("provider:Setup:saveProvider:dhcp.ReserveGlobal(): %s", err.Error())
		return err
	}

	// retrieve the Hosts known ip
	hIP, err := provider.HostIP()
	if err != nil {
		lumber.Error("provider:Setup:saveProvider:provider.HostIP(): %s", err.Error())
		return err
	}

	mProvider.HostIP = hIP
	mProvider.MountIP = ip.String()

	return mProvider.Save()
}

func (setup Setup) SetDefaultIP() error {
	mProvider, _ := models.LoadProvider()

	if err := provider.AddIP(mProvider.MountIP); err != nil {
		lumber.Error("provider:Setup:SetDefaultIP:provider.AddIP(%s): %s", mProvider.MountIP, err.Error())
		return err
	}

	return provider.SetDefaultIP(mProvider.MountIP)
}
