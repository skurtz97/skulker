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