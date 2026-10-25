# Installer checks

What the Windows installer was verified to do, and how. Run after changing
`mygo.json`, the icon, or the release workflow.

## Build and install

```sh
go tool mygo build -platform windows/amd64 -o check
start "check\windows-amd64\Muo Setup 0.1.0 amd64.exe" /S
```

`/S` is a silent install, which skips the finish page and so creates no desktop
shortcut. The Start Menu shortcut is always made.

## What to check

**No terminal.** The executable is a GUI app, not a console one. A console app
opens a terminal window alongside it, which on a reader looks like a bug.

```powershell
# Subsystem 2 is WINDOWS_GUI, 3 is WINDOWS_CUI.
python -c "
import struct
data = open(r'%LOCALAPPDATA%\Programs\Muo\Muo.exe','rb').read()
pe = struct.unpack_from('<I', data, 0x3c)[0]
print(struct.unpack_from('<H', data, pe + 0x5c)[0])
"
```

`mygo build` passes `-H=windowsgui`, so a packaged app reports 2. A plain
`go build` reports 3 and does open a terminal; that is expected, and is why the
release builds go through `mygo build`.

**The icon, in four places.** All four come from `resources/icon.png`:

| Where | How to see it |
| --- | --- |
| The executable | `ExtractAssociatedIcon` on `Muo.exe` |
| The installer | `ExtractAssociatedIcon` on `Muo Setup …exe` |
| The Start Menu shortcut | Properties, Change Icon: it points at the exe, index 0 |
| The taskbar button | Run the app and look |

**The uninstall entry.** Settings, Apps, Installed apps shows "Muo 0.1.0" with the
icon and the right location.

```powershell
Get-ItemProperty "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\com.muo.reader"
```

**Per user, no admin prompt.** It installs to `%LOCALAPPDATA%\Programs\Muo` and
writes `HKCU` only, so it needs no elevation.

**File associations do not hijack.** `.md` and `.pdf` are registered under
`OpenWithProgids`, which adds Myu to the "Open with" list without taking the
extension from whatever the reader already uses.

```powershell
Get-ItemProperty "HKCU:\Software\Classes\.md\OpenWithProgids"
# com.muo.reader.md should appear beside the reader's own editor
```

## Uninstall

```powershell
& "$env:LOCALAPPDATA\Programs\Muo\Uninstall.exe" /S
```

Removes the program directory, the Start Menu shortcut, the uninstall registry
key and the file associations. The reader's own data, in
`%APPDATA%\muo-dev` or `%APPDATA%\Muo`, is left alone, which is right: their
place in a document should survive a reinstall.
