package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	"github.com/tranquil-tr0/get/internal/manager"
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
	pm = manager.NewPackageManager(metadataPath)

	qt.NewQApplication(os.Args)

	window := qt.NewQMainWindow(nil)
	window.SetWindowTitle("Get GUI")
	window.SetMinimumSize(qt.NewQSize2(800, 600))

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
		metadata, err := pm.GetPackageManagerMetadata()
		if err != nil {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Failed to load package metadata:\n%v", err))
			msgBox.Exec()
			return
		}
		if len(metadata.Packages) == 0 {
			packageList.AddItem("No packages installed.")
		} else {
			for pkgID, pkg := range metadata.Packages {
				itemText := fmt.Sprintf("%s (Version: %s)", pkgID, pkg.Version)
				packageList.AddItem(itemText)
			}
		}
	}

	installButton.OnClicked(func() {
		repoURL := repoInput.Text()
		if repoURL == "" {
			return
		}
		pkgID, err := tools.ParseRepoURL(repoURL)
		if err != nil {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Invalid repository URL:\n%v", err))
			msgBox.Exec()
			return
		}

		if err := pm.InstallWithOptions(pkgID, "", nil); err != nil {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Failed to install package:\n%v", err))
			msgBox.Exec()
		} else {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Information)
			msgBox.SetWindowTitle("Success")
			msgBox.SetText(fmt.Sprintf("Successfully installed %s", pkgID))
			msgBox.Exec()
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
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Failed to remove package:\n%v", err))
			msgBox.Exec()
		} else {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Information)
			msgBox.SetWindowTitle("Success")
			msgBox.SetText(fmt.Sprintf("Successfully removed %s", pkgID))
			msgBox.Exec()
			refreshPackageList()
		}
	})

	updateButton.OnClicked(func() {
		if err := pm.UpdateAllPackages(); err != nil {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Failed to check for updates:\n%v", err))
			msgBox.Exec()
		} else {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Information)
			msgBox.SetWindowTitle("Success")
			msgBox.SetText("Update check complete.")
			msgBox.Exec()
		}
	})

	upgradeButton.OnClicked(func() {
		if err := pm.UpgradeAllPackages(); err != nil {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Critical)
			msgBox.SetWindowTitle("Error")
			msgBox.SetText(fmt.Sprintf("Failed to upgrade packages:\n%v", err))
			msgBox.Exec()
		} else {
			msgBox := qt.NewQMessageBox(window.QWidget)
			msgBox.SetIcon(qt.QMessageBox__Information)
			msgBox.SetWindowTitle("Success")
			msgBox.SetText("All packages upgraded successfully.")
			msgBox.Exec()
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
