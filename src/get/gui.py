import sys
import os
from pathlib import Path
from PySide6.QtGui import QGuiApplication
from PySide6.QtQml import QQmlApplicationEngine
from PySide6.QtCore import QObject, Slot, Signal, Property, QStringListModel

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

def main():
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

    engine.load(qml_file)

    if not engine.rootObjects():
        sys.exit(-1)

    sys.exit(app.exec())
