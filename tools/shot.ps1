Get-Process muo-dev -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 500
Start-Process -FilePath 'Y:\Learn\Go\muo\muo-dev.exe' -ArgumentList 'Y:\Learn\Go\muo\sample.md'
Start-Sleep -Seconds 6
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$sig = '[DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);'
$w = Add-Type -MemberDefinition $sig -Name Win -Namespace Native -PassThru
$proc = Get-Process muo-dev -ErrorAction SilentlyContinue | Where-Object { $_.MainWindowHandle -ne 0 } | Select-Object -First 1
if ($proc) { $w::SetForegroundWindow($proc.MainWindowHandle) | Out-Null }
Start-Sleep -Seconds 2
$b = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$bmp = New-Object System.Drawing.Bitmap $b.Width, $b.Height
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen($b.Location, [System.Drawing.Point]::Empty, $b.Size)
$bmp.Save('Y:\Learn\Go\muo\live.png')
Write-Output "saved"
