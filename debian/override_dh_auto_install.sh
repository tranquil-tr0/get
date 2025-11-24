#!/bin/sh
set -eu
# Helper script run during override_dh_auto_install to prepare get and get-gui
# staging directories. Keeps complex shell logic out of debian/rules.

# Ensure get-gui staging exists so dh_install can find package paths
mkdir -p debian/get-gui/usr/bin
mkdir -p debian/get-gui/usr/share/get/qml

# Create placeholders in debian/tmp so dh_install can find files even when
# the build host did not produce GUI assets. These are harmless and will be
# overwritten if real assets are available.
mkdir -p debian/tmp/usr/bin
mkdir -p debian/tmp/usr/share/get/qml
# If pybuild didn't produce a launcher, create a simple Python module launcher
# that will invoke the package's `get.gui:main`. This requires the runtime
# dependency (PySide6) to be available on the target system.
if [ ! -f debian/tmp/usr/bin/get-gui ]; then
  printf '%s\n' '#!/bin/sh' 'exec /usr/bin/python3 -m get.gui "$$@"' > debian/tmp/usr/bin/get-gui || true
  chmod 755 debian/tmp/usr/bin/get-gui || true
fi
: > debian/tmp/usr/share/get/qml/placeholder || true

# Try moving QML assets from dh_auto_install output locations
if [ -e debian/get/usr/lib/python3/dist-packages/get/qml ]; then
  mkdir -p debian/get-gui/usr/share/get/qml
  # Prefer copying contents so we can flatten any nested qml/qml directory
  if [ -d debian/get/usr/lib/python3/dist-packages/get/qml/qml ]; then
    cp -a debian/get/usr/lib/python3/dist-packages/get/qml/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
  else
    cp -a debian/get/usr/lib/python3/dist-packages/get/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
  fi
  rm -rf debian/get/usr/lib/python3/dist-packages/get/qml || true
elif [ -e debian/tmp/usr/lib/python3/dist-packages/get/qml ]; then
  mkdir -p debian/get-gui/usr/share/get/qml
  if [ -d debian/tmp/usr/lib/python3/dist-packages/get/qml/qml ]; then
    cp -a debian/tmp/usr/lib/python3/dist-packages/get/qml/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
  else
    cp -a debian/tmp/usr/lib/python3/dist-packages/get/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
  fi
  rm -rf debian/tmp/usr/lib/python3/dist-packages/get/qml || true
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
    if [ -d debian/get/usr/lib/python3/dist-packages/get/qml/qml ]; then
      cp -a debian/get/usr/lib/python3/dist-packages/get/qml/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
    else
      cp -a debian/get/usr/lib/python3/dist-packages/get/qml/* debian/get-gui/usr/share/get/qml/ 2>/dev/null || true
    fi
    rm -rf debian/get/usr/lib/python3/dist-packages/get/qml || true
  fi
  mkdir -p debian/get/usr/bin
  printf '%s\n' '#!/bin/sh' 'exec /usr/bin/python3 -m get.cli "$$@"' > debian/get/usr/bin/get
  chmod 755 debian/get/usr/bin/get
fi

# Ensure any placeholder installed in debian/get-gui is safe
if [ -f debian/get-gui/usr/bin/get-gui ] && [ ! -x debian/get-gui/usr/bin/get-gui ]; then
  chmod 755 debian/get-gui/usr/bin/get-gui || true
fi

# Ensure the `get.gui` module is available in the get-gui package by copying
# `src/get/gui.py` into the get-gui Python package area under debian/tmp so
# `dh_install` can pick it up for the `get-gui` binary package.
mkdir -p debian/tmp/usr/lib/python3/dist-packages/get
if [ -f src/get/gui.py ]; then
  cp -a src/get/gui.py debian/tmp/usr/lib/python3/dist-packages/get/gui.py || true
fi

# Copy staged get-gui assets into debian/tmp so dh_install can find them
if [ -d debian/get-gui/usr/share/get/qml ]; then
  mkdir -p debian/tmp/usr/share/get/qml
  cp -a debian/get-gui/usr/share/get/qml/ debian/tmp/usr/share/get/qml/ 2>/dev/null || true
fi

if [ -f debian/get-gui/usr/bin/get-gui ]; then
  mkdir -p debian/tmp/usr/bin
  cp -a debian/get-gui/usr/bin/get-gui debian/tmp/usr/bin/get-gui 2>/dev/null || true
fi

exit 0
