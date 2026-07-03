import os
import re

root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
skip_dirs = {
    os.path.join(root, "engine", "benchmark"),
    os.path.join(root, "engine", "collector"),
    os.path.join(root, "internal", "repofs"),
}
skip_files = {
    "api.go",
    "types.go",
    "const.go",
    "doc.go",
    "api_test.go",
}

repl = [
    (r"internal\.FunctionFilter", "config.FunctionFilter"),
    (r"internal\.Config\b", "config.Config"),
    (r"internal\.CIConfig", "config.CIConfig"),
    (r"internal\.CITrackingConfig", "config.CITrackingConfig"),
    (r"internal\.LoadFromFile", "config.LoadFromFile"),
    (r"internal\.CreateTemplate", "config.CreateTemplate"),
    (r"internal\.ConfigFilename", "config.Filename"),
    (r"internal\.GlobalSign", "config.GlobalSign"),
    (r"internal\.BenchArgs", "config.AutoArgs"),
    (r"internal\.CollectionArgs", "config.CollectionArgs"),
    (r"internal\.PrintConfiguration", "config.PrintAutoConfiguration"),
    (r"internal\.MainDirOutput", "workspace.MainDirOutput"),
    (r"internal\.ProfileTextDir", "workspace.ProfileTextDir"),
    (r"internal\.ProfileBinDir", "workspace.ProfileBinDir"),
    (r"internal\.PermDir", "workspace.PermDir"),
    (r"internal\.PermFile", "workspace.PermFile"),
    (r"internal\.TextExtension", "workspace.TextExtension"),
    (r"internal\.ProfileArtifactExtension", "workspace.ProfileArtifactExtension"),
    (r"internal\.FunctionsDirSuffix", "workspace.FunctionsDirSuffix"),
    (r"internal\.CleanOrCreateTag", "workspace.CleanOrCreateTag"),
    (r"internal\.FindGoModuleRoot", "workspace.FindModuleRoot"),
    (r"internal\.InfoCollectionSuccess", "workspace.InfoCollectionSuccess"),
    (r"internal\.GoBinaryName", "workspace.GoBinaryName"),
    (r"internal\.GoTestSubcommand", "workspace.GoTestSubcommand"),
    (r"internal\.AUTOCMD", "cli.CmdAuto"),
    (r"internal\.MANUALCMD", "cli.CmdManual"),
]


def add_import(c: str, pkg: str) -> str:
    if pkg in c:
        return c
    needle = '"github.com/AlexsanderHamir/prof/internal/config"'
    if 'import (' in c:
        return c.replace("import (", f'import (\n\t"{pkg}"', 1)
    return c


def process_file(path: str) -> None:
    with open(path, "r", encoding="utf-8") as f:
        c = f.read()
    orig = c
    for pat, rep in repl:
        c = re.sub(pat, rep, c)
    if c == orig:
        return

    if "config." in c:
        c = add_import(c, "github.com/AlexsanderHamir/prof/internal/config")
    if "workspace." in c:
        c = add_import(c, "github.com/AlexsanderHamir/prof/internal/workspace")

    c = c.replace('"github.com/AlexsanderHamir/prof/internal"\n', "")
    c = c.replace('\t"github.com/AlexsanderHamir/prof/internal"\n', "")

    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(c)


for dirpath, dirnames, filenames in os.walk(root):
    if any(dirpath.startswith(s) for s in skip_dirs):
        continue
    for fn in filenames:
        if not fn.endswith(".go") or fn in skip_files:
            continue
        if os.path.join(dirpath, fn).endswith(tuple(
            os.path.join(root, "internal", x) for x in skip_files
        )):
            continue
        process_file(os.path.join(dirpath, fn))

print("migration complete")
