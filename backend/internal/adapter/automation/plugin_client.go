package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xiaoheiplay/internal/adapter/plugins"
	"xiaoheiplay/internal/usecase"
	pluginv1 "xiaoheiplay/plugin/v1"
)

type PluginInstanceClient struct {
	mgr        *plugins.Manager
	pluginID   string
	instanceID string
	timeout    time.Duration
}

func NewPluginInstanceClient(mgr *plugins.Manager, pluginID, instanceID string) *PluginInstanceClient {
	return &PluginInstanceClient{
		mgr:        mgr,
		pluginID:   strings.TrimSpace(pluginID),
		instanceID: strings.TrimSpace(instanceID),
		timeout:    12 * time.Second,
	}
}

func (c *PluginInstanceClient) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.timeout)
}

func (c *PluginInstanceClient) client(ctx context.Context) (pluginv1.AutomationServiceClient, error) {
	if c.mgr == nil {
		return nil, errors.New("plugin manager missing")
	}
	cli, _, err := c.mgr.GetAutomationClient(ctx, c.pluginID, c.instanceID)
	return cli, err
}

func mapUnimplemented(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	if st.Code() == codes.Unimplemented {
		msg := strings.TrimSpace(st.Message())
		if msg == "" {
			msg = "not supported"
		}
		return fmt.Errorf("%w: %s", usecase.ErrNotSupported, msg)
	}
	return err
}

func (c *PluginInstanceClient) CreateHost(ctx context.Context, req usecase.AutomationCreateHostRequest) (usecase.AutomationCreateHostResult, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return usecase.AutomationCreateHostResult{}, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.CreateInstance(cctx, &pluginv1.CreateInstanceRequest{
		LineId:        req.LineID,
		Os:            req.OS,
		Name:          req.HostName,
		Password:      req.SysPwd,
		VncPassword:   req.VNCPwd,
		ExpireAtUnix:  req.ExpireTime.Unix(),
		PortNum:       int32(req.PortNum),
		Cpu:           int32(req.CPU),
		MemoryGb:      int32(req.MemoryGB),
		DiskGb:        int32(req.DiskGB),
		BandwidthMbps: int32(req.Bandwidth),
	})
	if err != nil {
		return usecase.AutomationCreateHostResult{}, mapUnimplemented(err)
	}
	return usecase.AutomationCreateHostResult{HostID: resp.GetInstanceId(), Raw: map[string]any{"instance_id": resp.GetInstanceId()}}, nil
}

func (c *PluginInstanceClient) GetHostInfo(ctx context.Context, hostID int64) (usecase.AutomationHostInfo, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return usecase.AutomationHostInfo{}, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.GetInstance(cctx, &pluginv1.GetInstanceRequest{InstanceId: hostID})
	if err != nil {
		return usecase.AutomationHostInfo{}, mapUnimplemented(err)
	}
	inst := resp.GetInstance()
	var expire *time.Time
	if inst.GetExpireAtUnix() > 0 {
		t := time.Unix(inst.GetExpireAtUnix(), 0)
		expire = &t
	}
	return usecase.AutomationHostInfo{
		HostID:        inst.GetId(),
		HostName:      inst.GetName(),
		State:         int(inst.GetState()),
		CPU:           int(inst.GetCpu()),
		MemoryGB:      int(inst.GetMemoryGb()),
		DiskGB:        int(inst.GetDiskGb()),
		Bandwidth:     int(inst.GetBandwidthMbps()),
		PanelPassword: inst.GetPanelPassword(),
		VNCPassword:   inst.GetVncPassword(),
		OSPassword:    inst.GetOsPassword(),
		RemoteIP:      inst.GetRemoteIp(),
		ExpireAt:      expire,
	}, nil
}

func (c *PluginInstanceClient) ListHostSimple(ctx context.Context, searchTag string) ([]usecase.AutomationHostSimple, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListInstancesSimple(cctx, &pluginv1.ListInstancesSimpleRequest{SearchTag: strings.TrimSpace(searchTag)})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationHostSimple, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationHostSimple{ID: it.GetId(), HostName: it.GetName(), IP: it.GetIp()})
	}
	return out, nil
}

func (c *PluginInstanceClient) ElasticUpdate(ctx context.Context, req usecase.AutomationElasticUpdateRequest) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	pb := &pluginv1.ElasticUpdateRequest{InstanceId: req.HostID}
	if req.CPU != nil {
		pb.Cpu = ptrInt32(int32(*req.CPU))
	}
	if req.MemoryGB != nil {
		pb.MemoryGb = ptrInt32(int32(*req.MemoryGB))
	}
	if req.DiskGB != nil {
		pb.DiskGb = ptrInt32(int32(*req.DiskGB))
	}
	if req.Bandwidth != nil {
		pb.BandwidthMbps = ptrInt32(int32(*req.Bandwidth))
	}
	if req.PortNum != nil {
		pb.PortNum = ptrInt32(int32(*req.PortNum))
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.ElasticUpdate(cctx, pb)
	return mapUnimplemented(err)
}

func ptrInt32(v int32) *int32 { return &v }

func (c *PluginInstanceClient) RenewHost(ctx context.Context, hostID int64, nextDueDate time.Time) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Renew(cctx, &pluginv1.RenewRequest{InstanceId: hostID, NextDueAtUnix: nextDueDate.Unix()})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) LockHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Lock(cctx, &pluginv1.LockRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) UnlockHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Unlock(cctx, &pluginv1.UnlockRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) DeleteHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Destroy(cctx, &pluginv1.DestroyRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) StartHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Start(cctx, &pluginv1.StartRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ShutdownHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Shutdown(cctx, &pluginv1.ShutdownRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) RebootHost(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Reboot(cctx, &pluginv1.RebootRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ResetOS(ctx context.Context, hostID int64, templateID int64, password string) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.Rebuild(cctx, &pluginv1.RebuildRequest{InstanceId: hostID, ImageId: templateID, Password: password})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ResetOSPassword(ctx context.Context, hostID int64, password string) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.ResetPassword(cctx, &pluginv1.ResetPasswordRequest{InstanceId: hostID, Password: password})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ListSnapshots(ctx context.Context, hostID int64) ([]usecase.AutomationSnapshot, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListSnapshots(cctx, &pluginv1.ListSnapshotsRequest{InstanceId: hostID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationSnapshot, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationSnapshot{
			"id":              it.GetId(),
			"name":            it.GetName(),
			"created_at_unix": it.GetCreatedAtUnix(),
			"created_at":      time.Unix(it.GetCreatedAtUnix(), 0).Format(time.RFC3339),
		})
	}
	return out, nil
}

func (c *PluginInstanceClient) CreateSnapshot(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.CreateSnapshot(cctx, &pluginv1.CreateSnapshotRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) DeleteSnapshot(ctx context.Context, hostID int64, snapshotID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.DeleteSnapshot(cctx, &pluginv1.DeleteSnapshotRequest{InstanceId: hostID, SnapshotId: snapshotID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) RestoreSnapshot(ctx context.Context, hostID int64, snapshotID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.RestoreSnapshot(cctx, &pluginv1.RestoreSnapshotRequest{InstanceId: hostID, SnapshotId: snapshotID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ListBackups(ctx context.Context, hostID int64) ([]usecase.AutomationBackup, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListBackups(cctx, &pluginv1.ListBackupsRequest{InstanceId: hostID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationBackup, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationBackup{
			"id":              it.GetId(),
			"name":            it.GetName(),
			"created_at_unix": it.GetCreatedAtUnix(),
			"created_at":      time.Unix(it.GetCreatedAtUnix(), 0).Format(time.RFC3339),
		})
	}
	return out, nil
}

func (c *PluginInstanceClient) CreateBackup(ctx context.Context, hostID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.CreateBackup(cctx, &pluginv1.CreateBackupRequest{InstanceId: hostID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) DeleteBackup(ctx context.Context, hostID int64, backupID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.DeleteBackup(cctx, &pluginv1.DeleteBackupRequest{InstanceId: hostID, BackupId: backupID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) RestoreBackup(ctx context.Context, hostID int64, backupID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.RestoreBackup(cctx, &pluginv1.RestoreBackupRequest{InstanceId: hostID, BackupId: backupID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ListFirewallRules(ctx context.Context, hostID int64) ([]usecase.AutomationFirewallRule, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListFirewallRules(cctx, &pluginv1.ListFirewallRulesRequest{InstanceId: hostID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationFirewallRule, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationFirewallRule{
			"id":        it.GetId(),
			"direction": it.GetDirection(),
			"protocol":  it.GetProtocol(),
			"method":    it.GetMethod(),
			"port":      it.GetPort(),
			"ip":        it.GetIp(),
		})
	}
	return out, nil
}

func (c *PluginInstanceClient) AddFirewallRule(ctx context.Context, req usecase.AutomationFirewallRuleCreate) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.AddFirewallRule(cctx, &pluginv1.AddFirewallRuleRequest{
		InstanceId: req.HostID,
		Direction:  req.Direction,
		Protocol:   req.Protocol,
		Method:     req.Method,
		Port:       req.Port,
		Ip:         req.IP,
	})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) DeleteFirewallRule(ctx context.Context, hostID int64, ruleID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.DeleteFirewallRule(cctx, &pluginv1.DeleteFirewallRuleRequest{InstanceId: hostID, RuleId: ruleID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) ListPortMappings(ctx context.Context, hostID int64) ([]usecase.AutomationPortMapping, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListPortMappings(cctx, &pluginv1.ListPortMappingsRequest{InstanceId: hostID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationPortMapping, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationPortMapping{
			"id":    it.GetId(),
			"name":  it.GetName(),
			"sport": it.GetSport(),
			"dport": it.GetDport(),
		})
	}
	return out, nil
}

func (c *PluginInstanceClient) AddPortMapping(ctx context.Context, req usecase.AutomationPortMappingCreate) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.AddPortMapping(cctx, &pluginv1.AddPortMappingRequest{
		InstanceId: req.HostID,
		Name:       req.Name,
		Sport:      req.Sport,
		Dport:      req.Dport,
	})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) DeletePortMapping(ctx context.Context, hostID int64, mappingID int64) error {
	cli, err := c.client(ctx)
	if err != nil {
		return err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err = cli.DeletePortMapping(cctx, &pluginv1.DeletePortMappingRequest{InstanceId: hostID, MappingId: mappingID})
	return mapUnimplemented(err)
}

func (c *PluginInstanceClient) FindPortCandidates(ctx context.Context, hostID int64, keywords string) ([]int64, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.FindPortCandidates(cctx, &pluginv1.FindPortCandidatesRequest{InstanceId: hostID, Keywords: strings.TrimSpace(keywords)})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	return resp.GetPorts(), nil
}

func (c *PluginInstanceClient) GetPanelURL(ctx context.Context, hostName string, panelPassword string) (string, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return "", err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.GetPanelURL(cctx, &pluginv1.GetPanelURLRequest{InstanceName: hostName, PanelPassword: panelPassword})
	if err != nil {
		return "", mapUnimplemented(err)
	}
	return resp.GetUrl(), nil
}

func (c *PluginInstanceClient) ListAreas(ctx context.Context) ([]usecase.AutomationArea, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListAreas(cctx, &pluginv1.Empty{})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationArea, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationArea{ID: it.GetId(), Name: it.GetName(), State: int(it.GetState())})
	}
	return out, nil
}

func (c *PluginInstanceClient) ListImages(ctx context.Context, lineID int64) ([]usecase.AutomationImage, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListImages(cctx, &pluginv1.ListImagesRequest{LineId: lineID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationImage, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationImage{ImageID: it.GetId(), Name: it.GetName(), Type: it.GetType()})
	}
	return out, nil
}

func (c *PluginInstanceClient) ListLines(ctx context.Context) ([]usecase.AutomationLine, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListLines(cctx, &pluginv1.Empty{})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationLine, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationLine{ID: it.GetId(), Name: it.GetName(), AreaID: it.GetAreaId(), State: int(it.GetState())})
	}
	return out, nil
}

func (c *PluginInstanceClient) ListProducts(ctx context.Context, lineID int64) ([]usecase.AutomationProduct, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return nil, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.ListPackages(cctx, &pluginv1.ListPackagesRequest{LineId: lineID})
	if err != nil {
		return nil, mapUnimplemented(err)
	}
	out := make([]usecase.AutomationProduct, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		out = append(out, usecase.AutomationProduct{
			ID:        it.GetId(),
			Name:      it.GetName(),
			CPU:       int(it.GetCpu()),
			MemoryGB:  int(it.GetMemoryGb()),
			DiskGB:    int(it.GetDiskGb()),
			Bandwidth: int(it.GetBandwidthMbps()),
			Price:     it.GetMonthlyPrice(),
			PortNum:   int(it.GetPortNum()),
		})
	}
	return out, nil
}

func (c *PluginInstanceClient) GetMonitor(ctx context.Context, hostID int64) (usecase.AutomationMonitor, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return usecase.AutomationMonitor{}, err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.GetMonitor(cctx, &pluginv1.GetMonitorRequest{InstanceId: hostID})
	if err != nil {
		return usecase.AutomationMonitor{}, mapUnimplemented(err)
	}
	if strings.TrimSpace(resp.GetRawJson()) == "" {
		return usecase.AutomationMonitor{}, nil
	}
	var raw struct {
		StorageStats float64         `json:"StorageStats"`
		NetworkStats json.RawMessage `json:"NetworkStats"`
		CpuStats     float64         `json:"CpuStats"`
		MemoryStats  float64         `json:"MemoryStats"`
	}
	if err := json.Unmarshal([]byte(resp.GetRawJson()), &raw); err != nil {
		return usecase.AutomationMonitor{}, err
	}
	bytesIn, bytesOut := parseNetworkStats(raw.NetworkStats)
	return usecase.AutomationMonitor{
		CPUPercent:     int(math.Round(raw.CpuStats)),
		MemoryPercent:  int(math.Round(raw.MemoryStats)),
		StoragePercent: int(math.Round(raw.StorageStats)),
		BytesIn:        bytesIn,
		BytesOut:       bytesOut,
	}, nil
}

func (c *PluginInstanceClient) GetVNCURL(ctx context.Context, hostID int64) (string, error) {
	cli, err := c.client(ctx)
	if err != nil {
		return "", err
	}
	cctx, cancel := c.withTimeout(ctx)
	defer cancel()
	resp, err := cli.GetVNCURL(cctx, &pluginv1.GetVNCURLRequest{InstanceId: hostID})
	if err != nil {
		return "", mapUnimplemented(err)
	}
	return resp.GetUrl(), nil
}
