#!/bin/bash -x

echo "Running buildpkgs..."

# Define variables
PACKAGE_NAME="chperm"
VERSION="${2:-1.0.0}" 
echo "Version: $VERSION"
ARCHITECTURE="amd64" 
SOURCE_DIR="dist/usr"  
BUILD_DIR="build"
DEBIAN_DIR="$BUILD_DIR/debian"
RPM_DIR="$BUILD_DIR/rpm"
SHAR_DIR="$BUILD_DIR/shar"
MAN_DIR="${SOURCE_DIR}/share/man/man1/"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    rm -rf "$BUILD_DIR"
    sed -i "s/$VERSION/@@VERSION@@/g" "main.go"
}

# Create Debian package
create_deb() {
    echo "Creating Debian package..."

    # Create directory structure
    DEB_DIR="$DEBIAN_DIR/$PACKAGE_NAME-$VERSION"
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/local/bin"
    mkdir -p "$DEB_DIR/usr/local/etc"
    mkdir -p "$DEB_DIR/usr/share/man/man1"
    mkdir -p "$DEB_DIR/usr/local/share/rollback"

    # Copy files
    cp "$SOURCE_DIR/local/bin/chperm" "$DEB_DIR/usr/local/bin/"
    cp "$SOURCE_DIR/local/etc/chperm.conf" "$DEB_DIR/usr/local/etc/"
    cp "$MAN_DIR/chperm.1" "$DEB_DIR/usr/share/man/man1/"

    # Create control file
    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: $PACKAGE_NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCHITECTURE
Maintainer: Megalith <megalith@root.sh>
Description: A utility to apply Unix permissions
EOF

    # Build package
    dpkg-deb --build "$DEB_DIR" "$DEBIAN_DIR"
}

# Create RPM package
create_rpm() {
    echo "Creating RPM package..."

    # Define RPM build directories
    mkdir -p "$RPM_DIR/BUILD" "$RPM_DIR/RPMS" "$RPM_DIR/SOURCES" "$RPM_DIR/SPECS" "$RPM_DIR/SRPMS"

    # Copy files to SOURCES directory
    cp -dR $SOURCE_DIR/* "$RPM_DIR/BUILD/"

    # Create source tarball
    tar czf "$RPM_DIR/SOURCES/$PACKAGE_NAME-$VERSION.tar.gz" --transform 's,^,chperm-1.0.0/,' -C "$RPM_DIR/BUILD" .

    # Create spec file
    cat > "$RPM_DIR/SPECS/$PACKAGE_NAME.spec" << EOF
Name:           $PACKAGE_NAME
Version:        $VERSION
Release:        1%{?dist}
Summary:        A utility to apply Unix permissions

License:        GPLv3+
URL:            https://github.com/rickcollette/chperm
Source0:        %{name}-%{version}.tar.gz

%description
A utility to apply Unix permissions.

%prep
%setup -q

%build
# Add build steps if needed

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/usr/local/etc
mkdir -p %{buildroot}/usr/local/share/rollback
mkdir -p %{buildroot}/usr/share/man/man1
install -m 755 %{_builddir}/%{name}-%{version}/local/bin/chperm %{buildroot}/usr/local/bin/chperm
install -m 644 %{_builddir}/%{name}-%{version}/local/etc/chperm.conf %{buildroot}/usr/local/etc/chperm.conf
install -m 644 %{_builddir}/%{name}-%{version}/share/man/man1/chperm.1 %{buildroot}/usr/share/man/man1/chperm.1


%files
/usr/local/bin/chperm
/usr/local/etc/chperm.conf
/usr/share/man/man1/chperm.1.gz
/usr/local/share/rollback

%changelog
* Wed Nov 15 2023  Megalith <megalith@root.sh> - $VERSION-1
- Initial package
EOF

    # Build RPM package
    rpmbuild -ba "$RPM_DIR/SPECS/$PACKAGE_NAME.spec" --define "_topdir $(realpath $RPM_DIR)"

}

# Create shar file
create_shar() {
    echo "Creating shar file..."

    mkdir -p "$SHAR_DIR"

    # Create shar archive
    shar "$SOURCE_DIR/chperm" "$MAN_DIR/chperm.1" "$SOURCE_DIR/chperm.conf" > "$SHAR_DIR/$PACKAGE_NAME-$VERSION.shar"

    # Add extraction and permission logic
    cat >> "$SHAR_DIR/$PACKAGE_NAME-$VERSION.shar" << 'EOF'
# Extraction and permission logic
echo "Installing chperm..."
mkdir -p /usr/local/bin
mkdir -p /usr/local/etc
if ![[ -d /usr/share/man/man1 ]]
mkdir -p /usr/share/man/man1
fi
mkdir -p /usr/local/share/rollback
install -m 644 chperm.1 /usr/share/man/man1/chperm.1
install -m 755 chperm /usr/local/bin/chperm
install -m 644 chperm.conf /usr/local/etc/chperm.conf
echo "Done."
EOF
}

# Function to update version in files
update_vars_in_files() {
    sed -i "s/@@VERSION@@/$VERSION/g" "$MAN_DIR/chperm.1"
    sed -i "s/@@DATE@@/$(date +%Y-%m-%d)/g" "$MAN_DIR/chperm.1"
    sed -i "s/@@VERSION@@/$VERSION/g" "main.go"
}

# Main function
main() {
    cleanup
    update_vars_in_files
    create_deb
    create_rpm
    create_shar
    echo "Packaging complete."
}

main "$@"
