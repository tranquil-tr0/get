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
	installGroupBox.SetSizePolicy(*qt.NewQSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Fixed))
	layout.AddWidget(installGroupBox.QWidget)

	repoInput := qt.NewQLineEdit(nil)
	repoInput.SetPlaceholderText("Enter repository name or URL (e.g., tranquil-tr0/get, github.com/tranquil-tr0/get)")
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

	// Container for package widgets
	packageList := qt.NewQWidget(nil)
	packageListLayout := qt.NewQVBoxLayout(nil)
	packageListLayout.SetSpacing(5) // Fixed spacing between package widgets
	packageList.SetLayout(packageListLayout.QLayout)
	packageList.SetSizePolicy(*qt.NewQSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Fixed))
	listLayout.AddWidget(packageList)

	listLayout.AddStretch()

	// --- Actions Section ---
	actionsLayout := qt.NewQHBoxLayout(nil)
	layout.AddLayout(actionsLayout.QLayout)

	updateButton := qt.NewQPushButton(nil)
	updateButton.SetText("Check for Updates")
	actionsLayout.AddWidget(updateButton.QWidget)

	upgradeAllPackagesButton := qt.NewQPushButton(nil)
	upgradeAllPackagesButton.SetText("Upgrade All")
	actionsLayout.AddWidget(upgradeAllPackagesButton.QWidget)

	// --- Functionality ---

	var addPackageWidget func(pkgID string, pkg manager.PackageMetadata, updateVersion string)

	populatePackageList := func() {
		// Remove all widgets from the layout
		for packageListLayout.Count() > 0 {
			item := packageListLayout.TakeAt(0)
			if item != nil {
				w := item.Widget()
				if w != nil {
					w.DeleteLater()
				}
			}
		}

		sortedKeys, packages, pendingUpdates, err := pm.ListInstalledPackagesAndPendingUpdates()
		if err != nil {
			pm.Out.PrintError("Failed to load package metadata:\n%v", err)
			return
		}
		if len(sortedKeys) == 0 {
			label := qt.NewQLabel(nil)
			label.SetText("No packages installed.")
			packageListLayout.AddWidget(label.QWidget)
			return
		}

		// Show packages with pending updates at the top
		for _, pkgID := range sortedKeys {
			newVersion, hasUpdate := pendingUpdates[pkgID]
			pkg := packages[pkgID]
			if hasUpdate {
				addPackageWidget(pkgID, pkg, newVersion)
			}
		}
		// Show the rest (no pending updates)
		for _, pkgID := range sortedKeys {
			if _, hasUpdate := pendingUpdates[pkgID]; hasUpdate {
				continue
			}
			pkg := packages[pkgID]
			addPackageWidget(pkgID, pkg, "")
		}
	}

	upgradePackageButtonClick := func(pkgID string) {
		if pkgID == "" {
			return
		}
		err := pm.UpgradeSpecificPackage(context.Background(), pkgID)
		if err != nil {
			pm.Out.PrintError("Failed to upgrade %s:\n%v", pkgID, err)
		} else {
			pm.Out.PrintSuccess("Successfully upgraded %s", pkgID)
			populatePackageList()
		}
	}
	// Helper to add a package widget
	addPackageWidget = func(pkgID string, pkg manager.PackageMetadata, updateVersion string) {
		pkgWidget := qt.NewQWidget(nil)
		pkgWidget.SetObjectName(*qt.NewQAnyStringView3("packageCard"))
		pkgWidget.SetStyleSheet("#packageCard { background-color: palette(base); border: 2px solid palette(mid); border-radius: 10px; }")
		hLayout := qt.NewQHBoxLayout(nil)
		pkgWidget.SetLayout(hLayout.QLayout)

		// Inline label group for type, name, version
		labelRow := qt.NewQWidget(nil)
		labelLayout := qt.NewQHBoxLayout(nil)
		labelLayout.SetSpacing(10)
		labelRow.SetLayout(labelLayout.QLayout)
		labelRow.SetSizePolicy(*qt.NewQSizePolicy2(qt.QSizePolicy__Fixed, qt.QSizePolicy__Fixed))

		nameLabel := qt.NewQLabel(nil)
		nameLabel.SetText(pkgID)
		labelLayout.AddWidget(nameLabel.QWidget)

		versionLabel := qt.NewQLabel(nil)
		versionLabel.SetText(pkg.Version)
		versionLabel.SetStyleSheet("color: palette(dark);")
		labelLayout.AddWidget(versionLabel.QWidget)

		typeLabel := qt.NewQLabel(nil)
		typeLabel.SetText(fmt.Sprintf("(%s)", pkg.InstallType))
		typeLabel.SetStyleSheet("color: palette(mid);")
		labelLayout.AddWidget(typeLabel.QWidget)

		hLayout.AddWidget(labelRow)
		hLayout.AddStretch()

		// If update available, add button
		if updateVersion != "" {
			updateBtn := qt.NewQPushButton(nil)
			updateBtn.SetText(fmt.Sprintf("Upgrade to %s", updateVersion))
			updateBtn.SetSizePolicy(*qt.NewQSizePolicy2(qt.QSizePolicy__Fixed, qt.QSizePolicy__Fixed))
			updateBtn.OnClicked(func() {
				upgradePackageButtonClick(pkgID)
			})
			hLayout.AddWidget(updateBtn.QWidget)
			pkgWidget.SetStyleSheet("#packageCard { background-color: palette(base); border: 2px solid palette(dark); border-radius: 10px; }")
		}

		// Remove button
		removeBtn := qt.NewQPushButton(nil)
		removeBtn.SetText("Remove")
		removeBtn.SetSizePolicy(*qt.NewQSizePolicy2(qt.QSizePolicy__Fixed, qt.QSizePolicy__Fixed))
		removeBtn.OnClicked(func() {
			if err := pm.Remove(pkgID); err != nil {
				pm.Out.PrintError("Failed to remove package:\n%v", err)
			} else {
				pm.Out.PrintSuccess("Successfully removed %s", pkgID)
				populatePackageList()
			}
		})
		hLayout.AddWidget(removeBtn.QWidget)

		packageListLayout.AddWidget(pkgWidget)
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
				// User cancelled, don't show error
				return
			}
			pm.Out.PrintError("Failed to install package:\n%v", err)
		} else {
			pm.Out.PrintSuccess("Successfully installed %s", pkgID)
			repoInput.Clear()
			populatePackageList()
		}
	})

	updateButton.OnClicked(func() {
		pm.Out.PrintStatus("Checking for updates...")
		newUpdates, err := pm.UpdateAllPackages()
		if err != nil {
			pm.Out.PrintError("%v", err)
		}

		if len(newUpdates) == 0 {
			pm.Out.PrintInfo("No updates available.")
			return
		}

		var updateText strings.Builder
		updateText.WriteString(fmt.Sprintf("Found %d new updates:\n", len(newUpdates)))
		for pkgID, version := range newUpdates {
			updateText.WriteString(fmt.Sprintf("  %s: %s\n", pkgID, version))
		}
		pm.Out.PrintInfo(updateText.String())
		populatePackageList()
	})

	upgradeAllPackagesButton.OnClicked(func() {
		if err := pm.UpgradeAllPackages(context.Background()); err != nil {
			pm.Out.PrintError("Failed to upgrade packages:\n%v", err)
		} else {
			pm.Out.PrintSuccess("All packages upgraded successfully.")
			populatePackageList()
		}
	})

	// Initial load
	populatePackageList()

	window.Show()
	qt.QApplication_Exec()
}
