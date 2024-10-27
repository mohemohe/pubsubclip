# Maintainer: Your Name <youremail@example.com>
pkgname=pubsubclip
pkgver=git
pkgrel=1
pkgdesc="publish / subscribe clipboard via redis. supports text and image"
arch=('x86_64' 'aarch64')
url="https://github.com/mohemohe/pubsubclip"
license=('WTFPL')
depends=("wl-clipboard")
makedepends=('go' 'git')
install="archlinux/pubsubclip.install"
source=("git+https://github.com/mohemohe/pubsubclip.git#branch=develop")
sha256sums=('SKIP')

build() {
  cd "$srcdir/$pkgname"
  go build -o "${pkgname}"
}

package() {
  cd "$srcdir/$pkgname"
  install -Dm755 "${pkgname}" "$pkgdir/usr/bin/${pkgname}"

  # Systemd unit file installation
  install -Dm644 "$srcdir/$pkgname/archlinux/pubsubclip@.service" "$pkgdir/usr/lib/systemd/system/pubsubclip@.service"

  # Default environment file installation
  install -Dm644 "$srcdir/$pkgname/archlinux/pubsubclip.default" "$pkgdir/etc/default/pubsubclip"
}
