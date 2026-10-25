Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName System.Windows.Forms
$sig = @'
using System;
using System.Runtime.InteropServices;
public class MuoShot2 {
  [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
  [DllImport("user32.dll")] public static extern bool SetWindowPos(IntPtr hWnd, IntPtr after, int x, int y, int cx, int cy, uint flags);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool SetProcessDPIAware();
}
'@
Add-Type -TypeDefinition $sig
[MuoShot2]::SetProcessDPIAware() | Out-Null

$proc = Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.MainWindowTitle -like "*Muo*" -and $_.MainWindowHandle -ne 0 } | Select-Object -First 1
if (-not $proc) { Write-Output "app not running"; exit 1 }
$hwnd = $proc.MainWindowHandle

# Put it on the primary display, where it fits, so the whole window is captured.
$wa = [System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea
$w = [Math]::Min(1180, $wa.Width - 40)
$h = [Math]::Min(780, $wa.Height - 40)
[MuoShot2]::SetWindowPos($hwnd, [IntPtr]::Zero, $wa.X + 20, $wa.Y + 20, $w, $h, 0x0004) | Out-Null
Start-Sleep -Milliseconds 700
[MuoShot2]::SetForegroundWindow($hwnd) | Out-Null
Start-Sleep -Milliseconds 900

$r = New-Object MuoShot2+RECT
[MuoShot2]::GetWindowRect($hwnd, [ref]$r) | Out-Null
$w = $r.Right - $r.Left
$h = $r.Bottom - $r.Top
Write-Output "$($proc.ProcessName) at $($r.Left),$($r.Top) ${w}x${h}"

$bmp = New-Object System.Drawing.Bitmap $w, $h
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen($r.Left, $r.Top, 0, 0, (New-Object System.Drawing.Size $w, $h))
$bmp.Save('Y:\Learn\Go\muo\app.png')
$bmp.Dispose()
Write-Output "saved app.png"
