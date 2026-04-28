Name:           skulker
Version:        %{pkg_version}
Release:        1%{?dist}
Summary:        Interactive Fedora System Cleaner

License:        MIT
URL:            https://github.com/skurtz97/skulker
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  make

%description
An interactive system cleaner for performing routine maintenance on a Fedora system. 

%prep
%autosetup

%build
make build

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/%{_bindir}
install -p -m 755 %{name} %{buildroot}/%{_bindir}/%{name}

%files
%{_bindir}/%{name}
EOF

# 2. Add, commit, and push the newly populated file
git add packaging/skulker.spec
git commit -m "build: populate empty rpm spec file"
git push origin main

# 3. Clean the bad artifacts and run the release
rm -rf build/
make release VERSION=v0.1.0