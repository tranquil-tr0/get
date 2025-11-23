import sys
import os
from pathlib import Path
from PySide6.QtGui import QGuiApplication
from PySide6.QtQml import QQmlApplicationEngine
from PySide6.QtCore import QObject, Slot, Signal, Property, QStringListModel, QUrl, qInstallMessageHandler

from .manager import PackageManager

class GUIBackend(QObject):
    def __init__(self):
        super().__init__()
        self.pm = PackageManager()
        self._packages = []
        self.refresh_packages()

    packagesChanged = Signal()

    @Property("QVariantList", notify=packagesChanged)
    def packages(self):
        return self._packages
    
    @Property("QVariantList", notify=packagesChanged)
    def pendingUpdates(self):
        updates = []
        for repo, version in self.pm.pending_updates.items():
            updates.append({"name": repo, "newVersion": version})
        return updates

    @Slot()
    def refresh_packages(self):
        pkgs = self.pm.list_installed()
        self._packages = [
            {"name": k, "version": v["version"], "installedAt": v["installed_at"]}
            for k, v in pkgs.items()
        ]
        self.packagesChanged.emit()

    @Slot(str)
    def installPackage(self, repo):
        # This should ideally be threaded to not block UI
        print(f"Installing {repo}...")
        try:
            self.pm.install(repo)
            self.refresh_packages()
        except Exception as e:
            print(f"Error: {e}")

    @Slot(str)
    def removePackage(self, repo):
        print(f"Removing {repo}...")
        try:
            self.pm.remove(repo)
            self.refresh_packages()
        except Exception as e:
            print(f"Error: {e}")

    @Slot()
    def checkForUpdates(self):
        print("Checking for updates...")
        try:
            self.pm.update_all()
            self.refresh_packages()
        except Exception as e:
            print(f"Error checking updates: {e}")

    @Slot(str)
    def upgradePackage(self, repo):
        print(f"Upgrading {repo}...")
        try:
            # GUI interactive callback not implemented yet, relying on auto-detect or best match
            self.pm.upgrade_package(repo)
            self.refresh_packages()
        except Exception as e:
            print(f"Error upgrading {repo}: {e}")

    @Slot()
    def upgradeAll(self):
        print("Upgrading all...")
        try:
            for repo in list(self.pm.pending_updates.keys()):
                self.pm.upgrade_package(repo)
            self.refresh_packages()
        except Exception as e:
            print(f"Error upgrading all: {e}")

def main():
    # Enable debug output for QML
    qInstallMessageHandler(lambda msg_type, context, msg: print(f"QML: {msg}"))

    # Set environment variables to help locate QML modules
    qml_path = os.environ.get("QML2_IMPORT_PATH", "")
    kirigami_paths = [
        "/usr/lib/x86_64-linux-gnu/qt6/qml",
        "/usr/lib/x86_64-linux-gnu/qt5/qml",
        "/usr/lib/qt6/qml"
    ]
    for path in kirigami_paths:
        if os.path.exists(path):
            if qml_path:
                os.environ["QML2_IMPORT_PATH"] = f"{qml_path}:{path}"
            else:
                os.environ["QML2_IMPORT_PATH"] = path
            print(f"Added QML path: {path}")
            break

    app = QGuiApplication(sys.argv)
    engine = QQmlApplicationEngine()

    backend = GUIBackend()
    engine.rootContext().setContextProperty("backend", backend)

    # Load QML
    # Assuming Main.qml is in the same directory or accessible
    qml_file = Path(__file__).parent / "qml" / "Main.qml"
    if not qml_file.exists():
        print(f"Error: QML file not found at {qml_file}")
        sys.exit(1)

    engine.load(QUrl.fromLocalFile(str(qml_file)))

    if not engine.rootObjects():
        sys.exit(-1)

    sys.exit(app.exec())

if __name__ == "__main__":
    main()
