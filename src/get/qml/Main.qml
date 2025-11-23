import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import org.kde.kirigami as Kirigami

Kirigami.ApplicationWindow {
    id: root
    title: "Get Package Manager"
    width: 800
    height: 600

    globalDrawer: Kirigami.GlobalDrawer {
        isMenu: true
        actions: [
            Kirigami.Action {
                text: "Installed"
                icon.name: "view-list-details"
                onTriggered: pageStack.push(installedPage)
            },
            Kirigami.Action {
                text: "About"
                icon.name: "help-about"
                onTriggered: pageStack.push(aboutPage)
            }
        ]
    }

    pageStack.initialPage: installedPage

    Component {
        id: installedPage
        Kirigami.ScrollablePage {
            title: "Installed Packages"
            
            actions: [
                Kirigami.Action {
                    text: "Refresh"
                    icon.name: "view-refresh"
                    onTriggered: backend.refresh_packages()
                },
                Kirigami.Action {
                    text: "Install New"
                    icon.name: "list-add"
                    onTriggered: installDialog.open()
                }
            ]

            ListView {
                model: backend.packages
                delegate: Kirigami.SwipeListItem {
                    contentItem: ColumnLayout {
                        Label {
                            text: modelData.name
                            font.bold: true
                        }
                        Label {
                            text: "Version: " + modelData.version
                            opacity: 0.7
                        }
                    }
                    actions: [
                        Kirigami.Action {
                            text: "Remove"
                            icon.name: "edit-delete"
                            onTriggered: backend.removePackage(modelData.name)
                        }
                    ]
                }
            }
        }
    }

    Component {
        id: aboutPage
        Kirigami.Page {
            title: "About"
            Label {
                anchors.centerIn: parent
                text: "Get Package Manager\nVersion 0.5.0\n\nBuilt with Python & Kirigami"
                horizontalAlignment: Text.AlignHCenter
            }
        }
    }

    Dialog {
        id: installDialog
        title: "Install Package"
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        modal: true
        
        ColumnLayout {
            Label { text: "Repository (user/repo):" }
            TextField {
                id: repoInput
                placeholderText: "tranquil-tr0/get"
                Layout.fillWidth: true
            }
        }

        onAccepted: {
            if (repoInput.text !== "") {
                backend.installPackage(repoInput.text)
                repoInput.text = ""
            }
        }
    }
}
