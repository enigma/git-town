package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/git-town/git-town/v9/src/giturl"
	"github.com/git-town/git-town/v9/src/messages"
	"github.com/git-town/git-town/v9/src/stringslice"
)

// GitTown provides type-safe access to Git Town configuration settings
// stored in the local and global Git configuration.
type GitTown struct {
	Git
	originURLCache OriginURLCache
}

func NewGitTown(runner runner) *GitTown {
	return &GitTown{
		Git:            NewGit(runner),
		originURLCache: OriginURLCache{},
	}
}

type OriginURLCache map[string]*giturl.Parts

// AddToPerennialBranches registers the given branch names as perennial branches.
// The branches must exist.
func (gt *GitTown) AddToPerennialBranches(branches ...string) error {
	return gt.SetPerennialBranches(append(gt.PerennialBranches(), branches...))
}

func (gt *GitTown) BranchDurations() BranchDurations {
	return BranchDurations{
		MainBranch:        gt.MainBranch(),
		PerennialBranches: gt.PerennialBranches(),
	}
}

func (gt *GitTown) DeprecatedNewBranchPushFlagGlobal() string {
	return gt.globalConfigCache[DeprecatedNewBranchPushFlagKey]
}

func (gt *GitTown) DeprecatedNewBranchPushFlagLocal() string {
	return gt.localConfigCache[DeprecatedNewBranchPushFlagKey]
}

func (gt *GitTown) DeprecatedPushVerifyFlagGlobal() string {
	return gt.globalConfigCache[DeprecatedPushVerifyKey]
}

func (gt *GitTown) DeprecatedPushVerifyFlagLocal() string {
	return gt.localConfigCache[DeprecatedPushVerifyKey]
}

// GitAlias provides the currently set alias for the given Git Town command.
func (gt *GitTown) GitAlias(aliasType AliasType) string {
	return gt.GlobalConfigValue("alias." + string(aliasType))
}

// GitHubToken provides the content of the GitHub API token stored in the local or global Git Town configuration.
func (gt *GitTown) GitHubToken() string {
	return gt.LocalOrGlobalConfigValue(GithubTokenKey)
}

// GitLabToken provides the content of the GitLab API token stored in the local or global Git Town configuration.
func (gt *GitTown) GitLabToken() string {
	return gt.LocalOrGlobalConfigValue(GitlabTokenKey)
}

// GiteaToken provides the content of the Gitea API token stored in the local or global Git Town configuration.
func (gt *GitTown) GiteaToken() string {
	return gt.LocalOrGlobalConfigValue(GiteaTokenKey)
}

// HasBranchInformation indicates whether this configuration contains any branch hierarchy entries.
func (gt *GitTown) HasBranchInformation() bool {
	for key := range gt.localConfigCache {
		if strings.HasPrefix(key, "git-town-branch.") {
			return true
		}
	}
	return false
}

// HostingServiceName provides the name of the code hosting connector to use.
func (gt *GitTown) HostingServiceName() string {
	return gt.LocalOrGlobalConfigValue(CodeHostingDriverKey)
}

// HostingService provides the type-safe name of the code hosting connector to use.
// This function caches its result and can be queried repeatedly.
func (gt *GitTown) HostingService() (HostingService, error) {
	return NewHostingService(gt.HostingServiceName())
}

// IsMainBranch indicates whether the branch with the given name
// is the main branch of the repository.
func (gt *GitTown) IsMainBranch(branch string) bool {
	return branch == gt.MainBranch()
}

// IsOffline indicates whether Git Town is currently in offline mode.
func (gt *GitTown) IsOffline() (bool, error) {
	config := gt.GlobalConfigValue(OfflineKey)
	if config == "" {
		return false, nil
	}
	result, err := ParseBool(config)
	if err != nil {
		return false, fmt.Errorf(messages.ValueInvalid, OfflineKey, config)
	}
	return result, nil
}

// Lineage provides the configured ancestry information for this Git repo.
func (gt *GitTown) Lineage() Lineage {
	lineage := Lineage{}
	for _, key := range gt.LocalConfigKeysMatching(`^git-town-branch\..*\.parent$`) {
		child := strings.TrimSuffix(strings.TrimPrefix(key, "git-town-branch."), ".parent")
		parent := gt.LocalConfigValue(key)
		lineage[child] = parent
	}
	return lineage
}

// MainBranch provides the name of the main branch.
func (gt *GitTown) MainBranch() string {
	return gt.LocalOrGlobalConfigValue(MainBranchKey)
}

// MainBranch provides the name of the main branch, or the given default value if none is configured.
func (gt *GitTown) MainBranchOr(defaultValue string) string {
	configured := gt.LocalOrGlobalConfigValue(MainBranchKey)
	if configured != "" {
		return configured
	}
	return defaultValue
}

// OriginOverride provides the override for the origin hostname from the Git Town configuration.
func (gt *GitTown) OriginOverride() string {
	return gt.LocalConfigValue(CodeHostingOriginHostnameKey)
}

// OriginURLString provides the URL for the "origin" remote.
// Tests can stub this through the GIT_TOWN_REMOTE environment variable.
func (gt *GitTown) OriginURLString() string {
	remote := os.Getenv("GIT_TOWN_REMOTE")
	if remote != "" {
		return remote
	}
	output, _ := gt.QueryTrim("git", "remote", "get-url", OriginRemote)
	return output
}

// OriginURL provides the URL for the "origin" remote.
// Tests can stub this through the GIT_TOWN_REMOTE environment variable.
// Caches its result so can be called repeatedly.
func (gt *GitTown) OriginURL() *giturl.Parts {
	text := gt.OriginURLString()
	if text == "" {
		return nil
	}
	return DetermineOriginURL(text, gt.OriginOverride(), gt.originURLCache)
}

func DetermineOriginURL(originURL, originOverride string, originURLCache OriginURLCache) *giturl.Parts {
	cached, has := originURLCache[originURL]
	if has {
		return cached
	}
	url := giturl.Parse(originURL)
	if originOverride != "" {
		url.Host = originOverride
	}
	originURLCache[originURL] = url
	return url
}

// PerennialBranches returns all branches that are marked as perennial.
func (gt *GitTown) PerennialBranches() []string {
	result := gt.LocalOrGlobalConfigValue(PerennialBranchesKey)
	if result == "" {
		return []string{}
	}
	return strings.Split(result, " ")
}

// PullBranchStrategy provides the currently configured pull branch strategy.
func (gt *GitTown) PullBranchStrategy() (PullBranchStrategy, error) {
	text := gt.LocalOrGlobalConfigValue(PullBranchStrategyKey)
	return NewPullBranchStrategy(text)
}

// PushHook provides the currently configured push-hook setting.
func (gt *GitTown) PushHook() (bool, error) {
	err := gt.updateDeprecatedSetting(DeprecatedPushVerifyKey, PushHookKey)
	if err != nil {
		return false, err
	}
	setting := gt.LocalOrGlobalConfigValue(PushHookKey)
	if setting == "" {
		return true, nil
	}
	result, err := ParseBool(setting)
	if err != nil {
		return false, fmt.Errorf(messages.ValueInvalid, PushHookKey, setting)
	}
	return result, nil
}

// PushHook provides the currently configured push-hook setting.
func (gt *GitTown) PushHookGlobal() (bool, error) {
	err := gt.updateDeprecatedGlobalSetting(DeprecatedPushVerifyKey, PushHookKey)
	if err != nil {
		return false, err
	}
	setting := gt.GlobalConfigValue(PushHookKey)
	if setting == "" {
		return true, nil
	}
	result, err := ParseBool(setting)
	if err != nil {
		return false, fmt.Errorf(messages.ValueGlobalInvalid, PushHookKey, setting)
	}
	return result, nil
}

// RemoveFromPerennialBranches removes the given branch as a perennial branch.
func (gt *GitTown) RemoveFromPerennialBranches(branch string) error {
	return gt.SetPerennialBranches(stringslice.Remove(gt.PerennialBranches(), branch))
}

// RemoveLocalGitConfiguration removes all Git Town configuration.
func (gt *GitTown) RemoveLocalGitConfiguration() error {
	err := gt.Run("git", "config", "--remove-section", "git-town")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 128 {
				// Git returns exit code 128 when trying to delete a non-existing config section.
				// This is not an error condition in this workflow so we can ignore it here.
				return nil
			}
		}
		return fmt.Errorf(messages.ConfigRemoveError, err)
	}
	return nil
}

// RemoveMainBranchConfiguration removes the configuration entry for the main branch name.
func (gt *GitTown) RemoveMainBranchConfiguration() error {
	return gt.RemoveLocalConfigValue(MainBranchKey)
}

// RemoveParent removes the parent branch entry for the given branch
// from the Git configuration.
func (gt *GitTown) RemoveParent(branch string) error {
	return gt.RemoveLocalConfigValue("git-town-branch." + branch + ".parent")
}

// RemovePerennialBranchConfiguration removes the configuration entry for the perennial branches.
func (gt *GitTown) RemovePerennialBranchConfiguration() error {
	return gt.RemoveLocalConfigValue(PerennialBranchesKey)
}

// SetCodeHostingDriver sets the "github.code-hosting-driver" setting.
func (gt *GitTown) SetCodeHostingDriver(value string) error {
	gt.localConfigCache[CodeHostingDriverKey] = value
	err := gt.Run("git", "config", CodeHostingDriverKey, value)
	return err
}

// SetCodeHostingOriginHostname sets the "github.code-hosting-driver" setting.
func (gt *GitTown) SetCodeHostingOriginHostname(value string) error {
	gt.localConfigCache[CodeHostingOriginHostnameKey] = value
	err := gt.Run("git", "config", CodeHostingOriginHostnameKey, value)
	return err
}

// SetColorUI configures whether Git output contains color codes.
func (gt *GitTown) SetColorUI(value string) error {
	err := gt.Run("git", "config", "color.ui", value)
	return err
}

// SetMainBranch marks the given branch as the main branch
// in the Git Town configuration.
func (gt *GitTown) SetMainBranch(branch string) error {
	err := gt.SetLocalConfigValue(MainBranchKey, branch)
	return err
}

// SetNewBranchPush updates whether the current repository is configured to push
// freshly created branches to origin.
func (gt *GitTown) SetNewBranchPush(value bool, global bool) error {
	setting := strconv.FormatBool(value)
	if global {
		_, err := gt.SetGlobalConfigValue(PushNewBranchesKey, setting)
		return err
	}
	err := gt.SetLocalConfigValue(PushNewBranchesKey, setting)
	return err
}

// SetOffline updates whether Git Town is in offline mode.
func (gt *GitTown) SetOffline(value bool) error {
	_, err := gt.SetGlobalConfigValue(OfflineKey, strconv.FormatBool(value))
	return err
}

// SetParent marks the given branch as the direct parent of the other given branch
// in the Git Town configuration.
func (gt *GitTown) SetParent(branch, parentBranch string) error {
	err := gt.SetLocalConfigValue("git-town-branch."+branch+".parent", parentBranch)
	return err
}

// SetPerennialBranches marks the given branches as perennial branches.
func (gt *GitTown) SetPerennialBranches(branch []string) error {
	err := gt.SetLocalConfigValue(PerennialBranchesKey, strings.Join(branch, " "))
	return err
}

// SetPullBranchStrategy updates the configured pull branch strategy.
func (gt *GitTown) SetPullBranchStrategy(strategy PullBranchStrategy) error {
	err := gt.SetLocalConfigValue(PullBranchStrategyKey, string(strategy))
	return err
}

// SetPushHookLocally updates the configured pull branch strategy.
func (gt *GitTown) SetPushHookLocally(value bool) error {
	err := gt.SetLocalConfigValue(PushHookKey, strconv.FormatBool(value))
	return err
}

// SetPushHook updates the configured pull branch strategy.
func (gt *GitTown) SetPushHookGlobally(value bool) error {
	_, err := gt.SetGlobalConfigValue(PushHookKey, strconv.FormatBool(value))
	return err
}

// SetShouldShipDeleteRemoteBranch updates the configured pull branch strategy.
func (gt *GitTown) SetShouldShipDeleteRemoteBranch(value bool) error {
	err := gt.SetLocalConfigValue(ShipDeleteRemoteBranchKey, strconv.FormatBool(value))
	return err
}

// SetShouldSyncUpstream updates the configured pull branch strategy.
func (gt *GitTown) SetShouldSyncUpstream(value bool) error {
	err := gt.SetLocalConfigValue(SyncUpstreamKey, strconv.FormatBool(value))
	return err
}

func (gt *GitTown) SetSyncStrategy(value SyncStrategy) error {
	err := gt.SetLocalConfigValue(SyncStrategyKey, string(value))
	return err
}

func (gt *GitTown) SetSyncStrategyGlobal(value SyncStrategy) error {
	_, err := gt.SetGlobalConfigValue(SyncStrategyKey, string(value))
	return err
}

// SetTestOrigin sets the origin to be used for testing.
func (gt *GitTown) SetTestOrigin(value string) error {
	err := gt.SetLocalConfigValue(TestingRemoteURLKey, value)
	return err
}

// ShouldNewBranchPush indicates whether the current repository is configured to push
// freshly created branches up to origin.
func (gt *GitTown) ShouldNewBranchPush() (bool, error) {
	err := gt.updateDeprecatedSetting(DeprecatedNewBranchPushFlagKey, PushNewBranchesKey)
	if err != nil {
		return false, err
	}
	config := gt.LocalOrGlobalConfigValue(PushNewBranchesKey)
	if config == "" {
		return false, nil
	}
	value, err := ParseBool(config)
	if err != nil {
		return false, fmt.Errorf(messages.ValueInvalid, PushNewBranchesKey, config)
	}
	return value, nil
}

// ShouldNewBranchPushGlobal indictes whether the global configuration requires to push
// freshly created branches to origin.
func (gt *GitTown) ShouldNewBranchPushGlobal() (bool, error) {
	err := gt.updateDeprecatedGlobalSetting(DeprecatedNewBranchPushFlagKey, PushNewBranchesKey)
	if err != nil {
		return false, err
	}
	config := gt.GlobalConfigValue(PushNewBranchesKey)
	if config == "" {
		return false, nil
	}
	return ParseBool(config)
}

// ShouldShipDeleteOriginBranch indicates whether to delete the remote branch after shipping.
func (gt *GitTown) ShouldShipDeleteOriginBranch() (bool, error) {
	setting := gt.LocalOrGlobalConfigValue(ShipDeleteRemoteBranchKey)
	if setting == "" {
		return true, nil
	}
	result, err := strconv.ParseBool(setting)
	if err != nil {
		return true, fmt.Errorf(messages.ValueInvalid, ShipDeleteRemoteBranchKey, setting)
	}
	return result, nil
}

// ShouldSyncUpstream indicates whether this repo should sync with its upstream.
func (gt *GitTown) ShouldSyncUpstream() (bool, error) {
	text := gt.LocalOrGlobalConfigValue(SyncUpstreamKey)
	if text == "" {
		return true, nil
	}
	return ParseBool(text)
}

func (gt *GitTown) SyncStrategy() (SyncStrategy, error) {
	text := gt.LocalOrGlobalConfigValue(SyncStrategyKey)
	return ToSyncStrategy(text)
}

func (gt *GitTown) SyncStrategyGlobal() (SyncStrategy, error) {
	setting := gt.GlobalConfigValue(SyncStrategyKey)
	return ToSyncStrategy(setting)
}

func (gt *GitTown) updateDeprecatedSetting(deprecatedKey, newKey string) error {
	err := gt.updateDeprecatedLocalSetting(deprecatedKey, newKey)
	if err != nil {
		return err
	}
	return gt.updateDeprecatedGlobalSetting(deprecatedKey, newKey)
}

func (gt *GitTown) updateDeprecatedGlobalSetting(deprecatedKey, newKey string) error {
	deprecatedSetting := gt.GlobalConfigValue(deprecatedKey)
	if deprecatedSetting != "" {
		fmt.Printf("I found the deprecated global setting %q.\n", deprecatedKey)
		fmt.Printf("I am upgrading this setting to the new format %q.\n", newKey)
		_, err := gt.RemoveGlobalConfigValue(deprecatedKey)
		if err != nil {
			return err
		}
		_, err = gt.SetGlobalConfigValue(newKey, deprecatedSetting)
		return err
	}
	return nil
}

func (gt *GitTown) updateDeprecatedLocalSetting(deprecatedKey, newKey string) error {
	deprecatedSetting := gt.LocalConfigValue(deprecatedKey)
	if deprecatedSetting != "" {
		fmt.Printf("I found the deprecated local setting %q.\n", deprecatedKey)
		fmt.Printf("I am upgrading this setting to the new format %q.\n", newKey)
		err := gt.RemoveLocalConfigValue(deprecatedKey)
		if err != nil {
			return err
		}
		err = gt.SetLocalConfigValue(newKey, deprecatedSetting)
		return err
	}
	return nil
}