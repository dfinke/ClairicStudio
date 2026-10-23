#Requires -Version 7.2
[CmdletBinding()]
param([string]$OutputDirectory = $PSScriptRoot)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing
New-Item -ItemType Directory -Path (Join-Path $OutputDirectory 'windows') -Force | Out-Null

# Render vector geometry at each native Windows icon size. No font dependency.
function New-StudioIconFrame([int]$Size) {
    $bitmap = [Drawing.Bitmap]::new($Size, $Size)
    $graphics = [Drawing.Graphics]::FromImage($bitmap)
    $graphics.SmoothingMode = [Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.Clear([Drawing.Color]::Transparent)
    $graphics.ScaleTransform($Size / 40.0, $Size / 40.0)

    $tile = [Drawing.Drawing2D.GraphicsPath]::new()
    $tile.AddArc(1, 1, 14, 14, 180, 90)
    $tile.AddArc(25, 1, 14, 14, 270, 90)
    $tile.AddArc(25, 25, 14, 14, 0, 90)
    $tile.AddArc(1, 25, 14, 14, 90, 90)
    $tile.CloseFigure()
    $background = [Drawing.SolidBrush]::new([Drawing.ColorTranslator]::FromHtml('#205bc4'))
    $graphics.FillPath($background, $tile)

    $paper = [Drawing.Drawing2D.GraphicsPath]::new()
    $paper.AddLines([Drawing.PointF[]]@(
        [Drawing.PointF]::new(10, 7), [Drawing.PointF]::new(23, 7),
        [Drawing.PointF]::new(29, 13), [Drawing.PointF]::new(29, 32),
        [Drawing.PointF]::new(10, 32), [Drawing.PointF]::new(10, 7)
    ))
    $white = [Drawing.Pen]::new([Drawing.ColorTranslator]::FromHtml('#f2f6ff'), 2.3)
    $white.LineJoin = [Drawing.Drawing2D.LineJoin]::Round
    $white.StartCap = $white.EndCap = [Drawing.Drawing2D.LineCap]::Round
    $graphics.DrawPath($white, $paper)
    $graphics.DrawLines($white, [Drawing.PointF[]]@(
        [Drawing.PointF]::new(23, 7), [Drawing.PointF]::new(23, 13),
        [Drawing.PointF]::new(29, 13)
    ))
    $graphics.DrawLine($white, 15, 19, 23, 19)
    $graphics.DrawLine($white, 15, 24, 21, 24)

    # The gold extraction badge remains visible even in a 16-pixel title bar.
    $gold = [Drawing.SolidBrush]::new([Drawing.ColorTranslator]::FromHtml('#ffd54f'))
    $graphics.FillEllipse($gold, 22, 22, 15, 15)
    $ink = [Drawing.Pen]::new([Drawing.ColorTranslator]::FromHtml('#142c55'), 2.2)
    $ink.StartCap = $ink.EndCap = [Drawing.Drawing2D.LineCap]::Round
    $graphics.DrawLine($ink, 26, 27, 33, 27)
    $graphics.DrawLine($ink, 26, 32, 30, 32)

    $memory = [IO.MemoryStream]::new()
    try {
        $bitmap.Save($memory, [Drawing.Imaging.ImageFormat]::Png)
        return ,$memory.ToArray()
    } finally {
        $memory.Dispose()
        $graphics.Dispose()
        $bitmap.Dispose()
        $tile.Dispose()
        $paper.Dispose()
        $background.Dispose()
        $white.Dispose()
        $gold.Dispose()
        $ink.Dispose()
    }
}

$sizes = @(16, 20, 24, 32, 40, 48, 64, 128, 256)
$frames = @($sizes | ForEach-Object { New-StudioIconFrame $_ })
[IO.File]::WriteAllBytes((Join-Path $OutputDirectory 'appicon.png'), $frames[-1])

$iconPath = Join-Path $OutputDirectory 'windows/icon.ico'
$stream = [IO.File]::Create($iconPath)
$writer = [IO.BinaryWriter]::new($stream)
try {
    $writer.Write([uint16]0)
    $writer.Write([uint16]1)
    $writer.Write([uint16]$sizes.Count)
    $offset = 6 + 16 * $sizes.Count
    for ($i = 0; $i -lt $sizes.Count; $i++) {
        $dimension = if ($sizes[$i] -eq 256) { 0 } else { $sizes[$i] }
        $writer.Write([byte]$dimension)
        $writer.Write([byte]$dimension)
        $writer.Write([byte]0)
        $writer.Write([byte]0)
        $writer.Write([uint16]1)
        $writer.Write([uint16]32)
        $writer.Write([uint32]$frames[$i].Length)
        $writer.Write([uint32]$offset)
        $offset += $frames[$i].Length
    }
    foreach ($frame in $frames) {
        $writer.Write([byte[]]$frame)
    }
} finally {
    $writer.Dispose()
}
Write-Host "Generated $iconPath and appicon.png"
