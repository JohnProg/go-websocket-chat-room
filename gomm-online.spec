%define BUILDTIME %(date +%Y_%m_%d_%H_%M)
%define REPONAME gomm-online
Name: gomm-online-trunk
Version: 1.0.0
Release: %{BUILDTIME}
Summary: part of qing.wps.cn, for the mutil-markdown-editor.

Group:		Application/Server
License:	GPL
URL:		http://qing_wps_cn/
Source0:	%{name}-%{version}.tar.gz
BuildRoot:	%_topdir/BUILDROOT
Prefix:		/opt/apps/web
BuildRequires: 

%description
wps.cn provide cooperative work

%prep


%build


%install
rm -rf %{buildroot}/
mkdir -p %{buildroot}/opt/apps/web
tar zxf %{SOURCE0} -C %{buildroot}/opt/apps/web
cd %{buildroot}/opt/apps/web
#mv %{name}-%{version} %{name}
cd %{REPONAME}


%clean
rm -rf %{buildroot}

%pre
mkdir -p /opt/apps/web

%post
cp /opt/apps/web/gomm-online/gomm-online.service /etc/init.d/gomm-online -f
chmod +x /etc/init.d/gomm-online

%preun

%postun
if [ $1 =0 ]; then
    rm -rf /opt/apps/web/gomm-online 1>/dev/null 2>&1
fi

%files
/opt/apps/web/%{REPONAME}

%doc

%changelog

