#!/usr/bin/env python3

import os
import subprocess
import sys
import shutil
import glob

def clean_build_artifacts():
    """Clean up all build artifacts and cache files."""
    # Clean up PyInstaller artifacts
    if os.path.exists("build"):
        shutil.rmtree("build")
    if os.path.exists("prof.spec"):
        os.remove("prof.spec")
    
    # Clean up __pycache__ directories
    for pycache in glob.glob("**/__pycache__", recursive=True):
        shutil.rmtree(pycache)
    for pyc in glob.glob("**/*.pyc", recursive=True):
        os.remove(pyc)

def build_binary():
    """Build a single binary using PyInstaller."""
    print("Building prof binary...")
    
    # Ensure we're in the right directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    # Clean up any previous builds
    clean_build_artifacts()
    
    # Run PyInstaller
    cmd = [
        "pyinstaller",
        "--onefile",  # Create a single binary
        "--name", "prof",  # Name of the output binary
        "--clean",  # Clean PyInstaller cache
        "--noconfirm",  # Replace existing spec file
        "--add-data", "templates:templates",  # Include templates directory
        "prof"  # The main script
    ]
    
    try:
        subprocess.run(cmd, check=True)
        print("\nBuild successful!")
        print(f"Binary location: {os.path.join(script_dir, 'dist', 'prof')}")
        
        # Make the binary executable
        binary_path = os.path.join(script_dir, "dist", "prof")
        os.chmod(binary_path, 0o755)
        
        # Clean up build artifacts after successful build
        clean_build_artifacts()
        
        print("\nTo install the binary, run:")
        print(f"cp {binary_path} ~/bin/")
        
    except subprocess.CalledProcessError as e:
        print(f"Error building binary: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    build_binary() 