#!/usr/bin/env python3
import os
import sys
import shutil
import subprocess
import time
import re
import json
from typing import List, Dict

def clean_directory(directory: str):
    """Delete all contents of a directory if it exists."""
    if os.path.exists(directory):
        try:
            # Remove all contents of the directory
            for item in os.listdir(directory):
                item_path = os.path.join(directory, item)
                if os.path.isfile(item_path):
                    os.remove(item_path)
                elif os.path.isdir(item_path):
                    shutil.rmtree(item_path)
            print(f"Cleaned directory: {directory}")
        except Exception as e:
            print(f"Error cleaning directory {directory}: {e}", file=sys.stderr)
            sys.exit(1)

def create_bench_directories(tag: str, benchmarks: List[str]):
    """Create 'bench' directory and a subdirectory with the tag name if they don't exist.
    Within the tag directory, create 'bin' and 'text' subdirectories with benchmark-specific subdirectories,
    and a description.txt file."""
    bench_dir = "bench"
    tag_dir = os.path.join(bench_dir, tag)
    bin_dir = os.path.join(tag_dir, "bin")
    text_dir = os.path.join(tag_dir, "text")
    description_file = os.path.join(tag_dir, "description.txt")
    
    try:
        # Create bench directory if it doesn't exist
        if not os.path.exists(bench_dir):
            os.makedirs(bench_dir)
            print(f"Created directory: {bench_dir}")
        else:
            print(f"Directory '{bench_dir}' already exists")
            
        # Clean and create tag subdirectory and its subdirectories
        if os.path.exists(tag_dir):
            print(f"Directory '{tag_dir}' already exists, cleaning it...")
            clean_directory(tag_dir)
        
        # Create the base directory structure
        os.makedirs(bin_dir)
        os.makedirs(text_dir)
        
        # Create benchmark-specific subdirectories
        for benchmark in benchmarks:
            os.makedirs(os.path.join(bin_dir, benchmark))
            os.makedirs(os.path.join(text_dir, benchmark))
        
        # Create empty description.txt file
        with open(description_file, 'w') as f:
            pass  # Create empty file
            
        print(f"Created directory structure: {tag_dir}")
        print(f"  - {bin_dir} (with benchmark subdirectories)")
        print(f"  - {text_dir} (with benchmark subdirectories)")
        print(f"  - {description_file}")
            
    except Exception as e:
        print(f"Error creating directories: {e}", file=sys.stderr)
        raise  # Re-raise the exception to be caught by main()

def parse_list_argument(arg: str) -> List[str]:
    """Parse a comma-separated string into a list, removing brackets if present."""
    # Remove brackets if present
    arg = arg.strip('[]')
    # Split by comma and strip whitespace
    return [item.strip() for item in arg.split(',')]

def wait_for_profile_file(profile_file: str, timeout: int = 5) -> bool:
    """Wait for a profile file to be written and become non-empty.
    Returns True if the file exists and is non-empty, False if timeout is reached."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        if os.path.exists(profile_file) and os.path.getsize(profile_file) > 0:
            return True
        time.sleep(0.1)
    return False

def run_benchmark(benchmark: str, profiles: List[str], count: int, tag: str):
    """Run a single benchmark with specified profiles and save output."""
    # Construct the benchmark command
    cmd = ["go", "test", "-run=^$", f"-bench=^{benchmark}$", "-benchmem", f"-count={count}"]
    
    # Add profile flags based on requested profiles
    profile_flags = {
        "cpu": "-cpuprofile=cpu.out",
        "memory": "-memprofile=memory.out",
        "mutex": "-mutexprofile=mutex.out",
        "trace": "-trace=trace.out"
    }
    
    for profile in profiles:
        if profile in profile_flags:
            cmd.append(profile_flags[profile])
    
    # Set up output directories
    tag_dir = os.path.join("bench", tag)
    text_dir = os.path.join(tag_dir, "text", benchmark)
    bin_dir = os.path.join(tag_dir, "bin", benchmark)
    
    # Set up output file (goes in benchmark-specific text directory)
    output_file = os.path.join(text_dir, f"{benchmark}.txt")
    
    # Run the benchmark and capture output
    with open(output_file, 'w') as f:
        process = subprocess.run(cmd, stdout=f, stderr=subprocess.STDOUT, text=True)
        
    if process.returncode != 0:
        with open(output_file, 'r') as f:
            error_output = f.read()
        raise RuntimeError(f"Benchmark {benchmark} failed:\n{error_output}")
        
    # Move profile files to the benchmark-specific bin directory if they exist
    for profile in profiles:
        if profile in profile_flags:
            profile_file = profile_flags[profile].split('=')[1]
            if os.path.exists(profile_file):
                # Wait for the profile file to be fully written
                if not wait_for_profile_file(profile_file):
                    print(f"Warning: Profile file {profile_file} was not fully written within timeout", file=sys.stderr)
                    continue
                    
                # Add benchmark name to profile file to avoid overwriting
                new_profile_file = f"{benchmark}_{profile}.out"
                shutil.move(profile_file, os.path.join(bin_dir, new_profile_file))
    
    # Look for any .test files in the current directory
    for item in os.listdir('.'):
        if item.endswith('.test'):
            # Wait for the test file to be fully written
            if not wait_for_profile_file(item):
                print(f"Warning: Test file {item} was not fully written within timeout", file=sys.stderr)
                continue
                
            # Add benchmark name to test file to avoid overwriting
            new_test_file = f"{benchmark}_{item}"
            shutil.move(item, os.path.join(bin_dir, new_test_file))
            print(f"Moved test file: {item} -> {new_test_file}")
    
    print(f"Completed benchmark: {benchmark}")

def process_profiles(benchmark: str, profiles: List[str], tag: str):
    """Process profile files using go tool pprof and generate PNG visualizations."""
    tag_dir = os.path.join("bench", tag)
    bin_dir = os.path.join(tag_dir, "bin", benchmark)
    text_dir = os.path.join(tag_dir, "text", benchmark)
    
    # Skip trace profile as it's not processed with pprof
    pprof_profiles = [p for p in profiles if p != "trace"]
    
    for profile in pprof_profiles:
        profile_file = os.path.join(bin_dir, f"{benchmark}_{profile}.out")
        if not os.path.exists(profile_file):
            print(f"Warning: Profile file not found: {profile_file}", file=sys.stderr)
            continue
            
        output_file = os.path.join(text_dir, f"{benchmark}_{profile}.txt")
        
        try:
            # Generate text profile analysis
            cmd = [
                "go", "tool", "pprof",
                "-nodecount=1000000",
                "-cum",
                "-edgefraction=0",
                "-nodefraction=0",
                "-top",
                profile_file
            ]
            
            # Run pprof and write output directly to file
            with open(output_file, 'w') as f:
                process = subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, text=True)
                
            if process.returncode != 0:
                raise RuntimeError(f"Error processing {profile} profile for {benchmark}: {process.stderr}")
                
            print(f"Processed {profile} profile for {benchmark}")
            
            # Generate PNG visualization in the appropriate profile_functions directory
            profile_functions_dir = os.path.join(tag_dir, f"{profile}_functions", benchmark)
            png_file = os.path.join(profile_functions_dir, f"{benchmark}_{profile}.png")
            png_cmd = ["go", "tool", "pprof", "-png", profile_file]
            
            with open(png_file, 'wb') as f:
                process = subprocess.run(png_cmd, stdout=f, stderr=subprocess.PIPE)
                
            if process.returncode != 0:
                print(f"Warning: Failed to generate PNG for {profile} profile of {benchmark}: {process.stderr.decode()}", file=sys.stderr)
            else:
                print(f"Generated PNG visualization for {profile} profile of {benchmark} in {profile_functions_dir}")
            
        except Exception as e:
            print(f"Error processing {profile} profile for {benchmark}: {e}", file=sys.stderr)
            raise  # Re-raise the exception to be caught by main()

def cleanup_tag_directory(tag: str):
    """Clean up the tag directory if it exists."""
    tag_dir = os.path.join("bench", tag)
    if os.path.exists(tag_dir):
        try:
            shutil.rmtree(tag_dir)
            print(f"Cleaned up tag directory: {tag_dir}")
        except Exception as e:
            print(f"Error cleaning up tag directory {tag_dir}: {e}", file=sys.stderr)

def create_profile_function_directories(tag: str, profiles: List[str], benchmarks: List[str]):
    """Create directories for profile function analysis."""
    tag_dir = os.path.join("bench", tag)
    
    # Skip trace profile
    pprof_profiles = [p for p in profiles if p != "trace"]
    
    for profile in pprof_profiles:
        profile_dir = os.path.join(tag_dir, f"{profile}_functions")
        os.makedirs(profile_dir, exist_ok=True)
        
        # Create benchmark subdirectories
        for benchmark in benchmarks:
            benchmark_dir = os.path.join(profile_dir, benchmark)
            os.makedirs(benchmark_dir, exist_ok=True)
            
    print("Created profile function directories")

def parse_benchmark_config(config_str: str) -> Dict[str, Dict[str, str]]:
    """Parse a JSON-like string containing benchmark-specific configurations.
    Example format: '{"BenchmarkGenPool":{"prefix":"github.com/AlexsanderHamir/GenPool","ignore":"func1,performWorkload"}}'"""
    try:
        # Replace single quotes with double quotes for valid JSON
        config_str = config_str.replace("'", '"')
        config = json.loads(config_str)
        
        # Validate the structure
        for benchmark, settings in config.items():
            if not isinstance(settings, dict):
                raise ValueError(f"Invalid settings format for benchmark {benchmark}")
            if "prefix" not in settings:
                raise ValueError(f"Missing 'prefix' for benchmark {benchmark}")
            if "ignore" in settings and not isinstance(settings["ignore"], str):
                raise ValueError(f"'ignore' must be a string for benchmark {benchmark}")
                
        return config
    except json.JSONDecodeError as e:
        raise ValueError(f"Invalid JSON format: {e}")
    except Exception as e:
        raise ValueError(f"Error parsing benchmark config: {e}")

def analyze_profile_functions(tag: str, profiles: List[str], benchmarks: List[str], 
                            benchmark_config: Dict[str, Dict[str, str]] = None):
    """Analyze profile files and extract function information from existing profile text files.
    If benchmark_config is None, analyze all functions for all benchmarks."""
    tag_dir = os.path.join("bench", tag)
    bin_dir = os.path.join(tag_dir, "bin")
    text_dir = os.path.join(tag_dir, "text")
    
    # Skip trace profile
    pprof_profiles = [p for p in profiles if p != "trace"]
    
    for profile in pprof_profiles:
        for benchmark in benchmarks:
            # Get configuration for this benchmark if available
            config = benchmark_config.get(benchmark, {}) if benchmark_config else {}
            function_prefix = config.get("prefix")
            ignore_functions = set(parse_list_argument(config.get("ignore", ""))) if config.get("ignore") else set()
            
            # Read the profile text file that contains the function list
            # Use benchmark-specific subdirectory
            profile_text_file = os.path.join(text_dir, benchmark, f"{benchmark}_{profile}.txt")
            if not os.path.exists(profile_text_file):
                print(f"Warning: Profile text file not found: {profile_text_file}", file=sys.stderr)
                continue
                
            try:
                # Read and process the profile text file
                functions = set()
                found_header = False
                
                with open(profile_text_file, 'r') as f:
                    for line in f:
                        line = line.strip()
                        
                        # Skip empty lines
                        if not line:
                            continue
                            
                        # Look for the header line
                        if "flat  flat%   sum%        cum   cum%" in line:
                            found_header = True
                            continue
                            
                        # Only process lines after the header
                        if not found_header:
                            continue
                            
                        # Split the line by whitespace and get the last column (function name)
                        parts = line.split()
                        if len(parts) < 6:  # Need at least 6 columns (flat, flat%, sum%, cum, cum%, function)
                            continue
                            
                        # Get the function name (everything after the last column)
                        func_name = " ".join(parts[5:])
                        
                        # If prefix is specified, only process matching functions
                        if function_prefix and function_prefix not in func_name:
                            continue
                            
                        # Extract just the function name after the last dot
                        match = re.search(r'\.([^.(]+)(?:\([^)]*\))?$', func_name)
                        if match:
                            func_name = match.group(1).strip().replace(" ", "")
                            if func_name and func_name not in ignore_functions:
                                functions.add(func_name)
                
                # For each function found, run detailed analysis
                for func in functions:
                    output_file = os.path.join(tag_dir, f"{profile}_functions", benchmark, f"{func}.txt")
                    # Use benchmark-specific subdirectory for profile file
                    profile_file = os.path.join(bin_dir, benchmark, f"{benchmark}_{profile}.out")
                    
                    # Run pprof with -list for the specific function
                    cmd = ["go", "tool", "pprof", f"-list={func}", profile_file]
                    
                    with open(output_file, 'w') as f:
                        process = subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, text=True)
                        
                    if process.returncode != 0:
                        print(f"Error analyzing function {func} in {profile_file}: {process.stderr}", file=sys.stderr)
                        continue
                        
                    print(f"Analyzed function {func} for {benchmark} ({profile})")
                    
            except Exception as e:
                print(f"Error processing {profile_text_file}: {e}", file=sys.stderr)
                continue 