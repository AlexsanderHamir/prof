$root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
if (-not (Test-Path (Join-Path $root "go.mod"))) { $root = "c:\Users\gomes\OneDrive\Documents\prof" }

Get-ChildItem -Path $root -Recurse -Filter '*.go' | Where-Object {
    $_.FullName -notmatch '\\engine\\benchmark\\' -and
    $_.FullName -notmatch '\\engine\\collector\\' -and
    $_.FullName -notmatch '\\internal\\repofs\\' -and
    $_.FullName -notmatch '\\internal\\api\.go' -and
    $_.FullName -notmatch '\\internal\\types\.go' -and
    $_.FullName -notmatch '\\internal\\const\.go' -and
    $_.FullName -notmatch '\\internal\\doc\.go' -and
    $_.FullName -notmatch '\\internal\\api_test\.go'
} | ForEach-Object {
    $c = Get-Content $_.FullName -Raw
    $orig = $c
    $replacements = @{
        'internal\.FunctionFilter' = 'config.FunctionFilter'
        'internal\.Config\b' = 'config.Config'
        'internal\.CIConfig' = 'config.CIConfig'
        'internal\.CITrackingConfig' = 'config.CITrackingConfig'
        'internal\.LoadFromFile' = 'config.LoadFromFile'
        'internal\.CreateTemplate' = 'config.CreateTemplate'
        'internal\.ConfigFilename' = 'config.Filename'
        'internal\.GlobalSign' = 'config.GlobalSign'
        'internal\.MainDirOutput' = 'workspace.MainDirOutput'
        'internal\.ProfileTextDir' = 'workspace.ProfileTextDir'
        'internal\.ProfileBinDir' = 'workspace.ProfileBinDir'
        'internal\.PermDir' = 'workspace.PermDir'
        'internal\.PermFile' = 'workspace.PermFile'
        'internal\.TextExtension' = 'workspace.TextExtension'
        'internal\.ProfileArtifactExtension' = 'workspace.ProfileArtifactExtension'
        'internal\.FunctionsDirSuffix' = 'workspace.FunctionsDirSuffix'
        'internal\.ToolDir' = 'workspace.ToolDir'
        'internal\.ToolNameBenchstat' = 'workspace.ToolNameBenchstat'
        'internal\.ToolNameQcachegrind' = 'workspace.ToolNameQcachegrind'
        'internal\.ToolsResultsSuffix' = 'workspace.ToolsResultsSuffix'
        'internal\.CleanOrCreateTag' = 'workspace.CleanOrCreateTag'
        'internal\.FindGoModuleRoot' = 'workspace.FindModuleRoot'
        'internal\.InfoCollectionSuccess' = 'workspace.InfoCollectionSuccess'
        'internal\.REGRESSION' = 'tracker.ChangeRegression'
        'internal\.IMPROVEMENT' = 'tracker.ChangeImprovement'
        'internal\.STABLE' = 'tracker.ChangeStable'
        'internal\.AUTOCMD' = 'cli.CmdAuto'
        'internal\.MANUALCMD' = 'cli.CmdManual'
        'internal\.BenchArgs' = 'config.AutoArgs'
        'internal\.CollectionArgs' = 'config.CollectionArgs'
        'internal\.PrintConfiguration' = 'config.PrintAutoConfiguration'
    }
    foreach ($k in $replacements.Keys) {
        $c = [regex]::Replace($c, $k, $replacements[$k])
    }
    if ($c -match 'config\.|workspace\.|tracker\.Change|cli\.Cmd') {
        if ($c -notmatch 'internal/config') {
            $c = $c -replace '"github.com/AlexsanderHamir/prof/internal"', "`"github.com/AlexsanderHamir/prof/internal/config`"`n`t`"github.com/AlexsanderHamir/prof/internal/workspace`""
        }
    }
    if ($c -ne $orig) { Set-Content $_.FullName $c -NoNewline }
}
