#!/usr/bin/env bash
set -euo pipefail

# Install and configure OCR/OpenCV toolchain for macOS CI.
# Mirrors docs/guides/ocr-integration.md guidance.

brew update
brew install pkg-config opencv tesseract leptonica

HOMEBREW_PREFIX="$(brew --prefix)"
LEPT_PREFIX="$(brew --prefix leptonica)"

# gosseract links with -llept while Homebrew may provide libleptonica only.
if [[ ! -f "${HOMEBREW_PREFIX}/lib/liblept.dylib" && -f "${LEPT_PREFIX}/lib/libleptonica.dylib" ]]; then
  ln -sf "${LEPT_PREFIX}/lib/libleptonica.dylib" "${HOMEBREW_PREFIX}/lib/liblept.dylib" || true
fi
if [[ -d /usr/local/lib && ! -f /usr/local/lib/liblept.dylib && -f "${LEPT_PREFIX}/lib/libleptonica.dylib" ]]; then
  sudo ln -sf "${LEPT_PREFIX}/lib/libleptonica.dylib" /usr/local/lib/liblept.dylib || true
fi

PKG_CONFIG_PATH_VALUE="${HOMEBREW_PREFIX}/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
CGO_CFLAGS_VALUE="-I${HOMEBREW_PREFIX}/include"
CGO_CPPFLAGS_VALUE="-I${HOMEBREW_PREFIX}/include"
CGO_CXXFLAGS_VALUE="-I${HOMEBREW_PREFIX}/include"
CGO_LDFLAGS_VALUE="-L${HOMEBREW_PREFIX}/lib -llept -ltesseract"

if [[ -n "${GITHUB_ENV:-}" ]]; then
  echo "PKG_CONFIG_PATH=${PKG_CONFIG_PATH_VALUE}" >> "$GITHUB_ENV"
  echo "CGO_CFLAGS=${CGO_CFLAGS_VALUE}" >> "$GITHUB_ENV"
  echo "CGO_CPPFLAGS=${CGO_CPPFLAGS_VALUE}" >> "$GITHUB_ENV"
  echo "CGO_CXXFLAGS=${CGO_CXXFLAGS_VALUE}" >> "$GITHUB_ENV"
  echo "CGO_LDFLAGS=${CGO_LDFLAGS_VALUE}" >> "$GITHUB_ENV"
else
  export PKG_CONFIG_PATH="${PKG_CONFIG_PATH_VALUE}"
  export CGO_CFLAGS="${CGO_CFLAGS_VALUE}"
  export CGO_CPPFLAGS="${CGO_CPPFLAGS_VALUE}"
  export CGO_CXXFLAGS="${CGO_CXXFLAGS_VALUE}"
  export CGO_LDFLAGS="${CGO_LDFLAGS_VALUE}"
fi

pkg-config --cflags lept tesseract
pkg-config --libs lept tesseract
