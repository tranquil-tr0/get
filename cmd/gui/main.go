package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	"github.com/tranquil-tr0/get/internal/manager"
	"github.com/tranquil-tr0/get/internal/output"
	"github.com/tranquil-tr0/get/internal/tools"
)

var pm *manager.PackageManager

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")

	qt.NewQApplication(os.Args)

	window := qt.NewQMainWindow(nil)
	window.SetWindowTitle("Get GUI")
	window.SetMinimumSize(qt.NewQSize2(800, 600))

	out := output.NewGUIOutput(window)
	pm = manager.NewPackageManager(metadataPath, out)

	centralWidget := qt.NewQWidget(nil)
	window.SetCentralWidget(centralWidget)

	layout := qt.NewQVBoxLayout(nil)
	centralWidget.SetLayout(layout.QLayout)

	// --- Install Section ---
	installGroupBox := qt.NewQGroupBox(nil)
	installGroupBox.SetTitle("Install Package")
	installLayout := qt.NewQHBoxLayout(nil)
	installGroupBox.SetLayout(installLayout.QLayout)
	layout.AddWidget(installGroupBox.QWidget)

	repoInput := qt.NewQLineEdit(nil)
	repoInput.SetPlaceholderText("Enter repository URL (e.g., tranquil-tr0/get)")
	installLayout.AddWidget(repoInput.QWidget)

	installButton := qt.NewQPushButton(nil)
	installButton.SetText("Install")
	installLayout.AddWidget(installButton.QWidget)

	// --- Installed Packages Section ---
	listGroupBox := qt.NewQGroupBox(nil)
	listGroupBox.SetTitle("Installed Packages")
	listLayout := qt.NewQVBoxLayout(nil)
	listGroupBox.SetLayout(listLayout.QLayout)
	layout.AddWidget(listGroupBox.QWidget)

	packageList := qt.NewQListWidget(nil)
	listLayout.AddWidget(packageList.QWidget)

	refreshButton := qt.NewQPushButton(nil)
	refreshButton.SetText("Refresh List")
	listLayout.AddWidget(refreshButton.QWidget)

	// --- Actions Section ---
	actionsLayout := qt.NewQHBoxLayout(nil)
	layout.AddLayout(actionsLayout.QLayout)

	removeButton := qt.NewQPushButton(nil)
	removeButton.SetText("Remove Selected")
	actionsLayout.AddWidget(removeButton.QWidget)

	updateButton := qt.NewQPushButton(nil)
	updateButton.SetText("Check for Updates")
	actionsLayout.AddWidget(updateButton.QWidget)

	upgradeButton := qt.NewQPushButton(nil)
	upgradeButton.SetText("Upgrade All")
	actionsLayout.AddWidget(upgradeButton.QWidget)

	// --- Functionality ---

	refreshPackageList := func() {
		packageList.Clear()
		sortedKeys, packages, pendingUpdates, err := pm.ListInstalledPackagesAndPendingUpdates()
		if err != nil {
			pm.Out.PrintError("Failed to load package metadata:\n%v", err)
			return
		}
		if len(sortedKeys) == 0 {
			packageList.AddItem("No packages installed.")
			return
		}

		// Show packages with pending updates at the top, in sorted order
		for _, pkgID := range sortedKeys {
			newVersion, hasUpdate := pendingUpdates[pkgID]
			pkg := packages[pkgID]
			if hasUpdate {
				itemText := fmt.Sprintf("%s v%s (%s) [Update Available to %s]", pkgID, pkg.Version, pkg.InstallType, newVersion)
				packageList.AddItem(itemText)
			}
		}
		// Show the rest (no pending updates), in sorted order
		for _, pkgID := range sortedKeys {
			if _, hasUpdate := pendingUpdates[pkgID]; hasUpdate {
				continue
			}
			pkg := packages[pkgID]
			itemText := fmt.Sprintf("%s v%s (%s)", pkgID, pkg.Version, pkg.InstallType)
			packageList.AddItem(itemText)
		}
	}

	installButton.OnClicked(func() {
		repoURL := repoInput.Text()
		if repoURL == "" {
			return
		}
		pkgID, err := tools.ParseRepoURL(repoURL)
		if err != nil {
			pm.Out.PrintError("Invalid repository URL:\n%v", err)
			return
		}

		if err := pm.InstallWithOptions(context.Background(), pkgID, "", nil); err != nil {
			if err == context.Canceled {
				// User cancelled, do nothing
				return
			}
			pm.Out.PrintError("Failed to install package:\n%v", err)
		} else {
			pm.Out.PrintSuccess("Successfully installed %s", pkgID)
			repoInput.Clear()
			refreshPackageList()
		}
	})

	removeButton.OnClicked(func() {
		currentItem := packageList.CurrentItem()
		if currentItem == nil {
			return
		}
		pkgText := currentItem.Text()
		pkgID := strings.Split(pkgText, " ")[0]

		if err := pm.Remove(pkgID); err != nil {
			pm.Out.PrintError("Failed to remove package:\n%v", err)
		} else {
			pm.Out.PrintSuccess("Successfully removed %s", pkgID)
			refreshPackageList()
		}
	})

	updateButton.OnClicked(func() {
		updates, err := pm.UpdateAllPackages()
		if err != nil {
			pm.Out.PrintError("Failed to check for updates:\n%v", err)
		}

		if len(updates) == 0 {
			pm.Out.PrintInfo("No updates available.")
			return
		}

		var updateText strings.Builder
		updateText.WriteString("Available updates:\n")
		for pkgID, version := range updates {
			updateText.WriteString(fmt.Sprintf("  %s: %s\n", pkgID, version))
		}
		pm.Out.PrintInfo(updateText.String())
		refreshPackageList()
	})

	upgradeButton.OnClicked(func() {
		if err := pm.UpgradeAllPackages(context.Background()); err != nil {
			pm.Out.PrintError("Failed to upgrade packages:\n%v", err)
		} else {
			pm.Out.PrintSuccess("All packages upgraded successfully.")
			refreshPackageList()
		}
	})

	refreshButton.OnClicked(func() {
		refreshPackageList()
	})

	// Initial load
	refreshPackageList()

	window.Show()
	qt.QApplication_Exec()
}
