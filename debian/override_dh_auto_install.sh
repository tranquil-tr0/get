#!/bin/sh
set -eu
# Helper script run during override_dh_auto_install to prepare get and get-gui
# staging directories. Keeps complex shell logic out of debian/rules.

# Ensure get-gui staging exists so dh_install can find package paths
mkdir -p debian/get-gui/usr/bin
mkdir -p debian/get-gui/usr/share/get/qml

# Try moving QML assets from dh_auto_install output locations
if [ -e debian/get/usr/lib/python3/dist-packages/get/qml ]; then
  mkdir -p debian/get-gui/usr/share/get/qml
  mv debian/get/usr/lib/python3/dist-packages/get/qml debian/get-gui/usr/share/get/qml || true
elif [ -e debian/tmp/usr/lib/python3/dist-packages/get/qml ]; then
  mkdir -p debian/get-gui/usr/share/get/qml
  mv debian/tmp/usr/lib/python3/dist-packages/get/qml debian/get-gui/usr/share/get/qml || true
fi

# Try moving GUI launcher
if [ -e debian/get/usr/bin/get-gui ]; then
  mkdir -p debian/get-gui/usr/bin
  mv debian/get/usr/bin/get-gui debian/get-gui/usr/bin/get-gui || true
elif [ -e debian/tmp/usr/bin/get-gui ]; then
  mkdir -p debian/get-gui/usr/bin
  mv debian/tmp/usr/bin/get-gui debian/get-gui/usr/bin/get-gui || true
fi

# Remove compiled caches that should not be packaged
if [ -d debian/get ]; then
  find debian/get -type d -name '__pycache__' -prune -exec rm -rf {} + || true
fi

# Remove any remaining GUI module files from the get package
rm -f debian/get/usr/lib/python3/dist-packages/get/gui.py || true

# Fallback: copy source tree into Debian layout if pybuild/dh_auto_install didn't
if [ ! -d debian/get/usr/lib/python3/dist-packages/get ]; then
  mkdir -p debian/get/usr/lib/python3/dist-packages
  cp -a src/get debian/get/usr/lib/python3/dist-packages/
  rm -f debian/get/usr/lib/python3/dist-packages/get/gui.py || true
  if [ -d debian/get/usr/lib/python3/dist-packages/get/qml ]; then
    mkdir -p debian/get-gui/usr/share/get/qml
    mv debian/get/usr/lib/python3/dist-packages/get/qml debian/get-gui/usr/share/get/qml || true
  fi
  mkdir -p debian/get/usr/bin
  printf '%s\n' '#!/bin/sh' 'exec /usr/bin/python3 -m get.cli "$$@"' > debian/get/usr/bin/get
  chmod 755 debian/get/usr/bin/get
fi

# Ensure any placeholder installed in debian/get-gui is safe
if [ -f debian/get-gui/usr/bin/get-gui ] && [ ! -x debian/get-gui/usr/bin/get-gui ]; then
  chmod 755 debian/get-gui/usr/bin/get-gui || true
fi

exit 0
