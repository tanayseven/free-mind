- [ ] Packaging
  - [ ] x86_64.rpm
  - [ ] aarch64.dmg
  - [ ] x64-setup.exe
  - [ ] x64_en-US.msi
  - [ ] x64.app.tar.gz
  - [ ] amd64.Appimage
  - [ ] amd64.deb
  - [ ] x64.dmg
  - [ ] aarch64.app.tar.gz


- [ ] Linux
```
["/etc/init.d/networking", "restart"],
["/etc/init.d/nscd", "restart"],
["/etc/rc.d/nscd", "restart"],
["/etc/rc.d/init.d/nscd", "restart"],
```

- [ ] MacOS
```
["dscacheutil", "-flushcache"]
```

- [ ] Windows
```
["ipconfig", "/flushdns"]
```
