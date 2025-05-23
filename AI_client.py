from pathlib import Path
from typing import Dict, List
from config_manager import ConfigManager, Config
import os

# Get the directory where the script is located
script_dir = Path(os.path.dirname(os.path.abspath(__file__)))

def read_profile_text_file(file_path: str) -> str:
    """Read and return the contents of a profile text file."""
    try:
        with open(file_path, 'r') as f:
            return f.read()
    except Exception as e:
        print(f"Error reading profile text file {file_path}: {e}")
        return ""

def get_benchmark_files(tag: str, benchmark_name: str, profile_type: str) -> Dict[str, str]:
    """Get the text file for a specific benchmark and profile type.
    
    Args:
        tag: The benchmark tag (e.g., "test1")
        benchmark_name: The name of the benchmark (e.g., "BenchmarkGenPool")
        profile_type: The type of profile (e.g., "cpu", "memory", "mutex")
        
    Returns:
        A dictionary containing:
        - text_content: The content of the profile text file
    """
    base_dir = Path("bench") / tag
    
    # Get text file
    text_file = base_dir / "text" / benchmark_name / f"{benchmark_name}_{profile_type}.txt"
    text_content = read_profile_text_file(str(text_file))
    
    return {
        "text_content": text_content
    }

def save_analysis(tag: str, benchmark_name: str, profile_type: str, analysis: str) -> None:
    """Save the AI analysis to a file in the specified directory structure.
    
    Args:
        tag: The benchmark tag
        benchmark_name: The name of the benchmark
        profile_type: The type of profile
        analysis: The analysis text to save
    """
    # Create the directory structure
    analysis_dir = Path("bench") / tag / "AI" / "generalistic" / benchmark_name
    analysis_dir.mkdir(parents=True, exist_ok=True)
    
    # Create the analysis file
    analysis_file = analysis_dir / f"generalistic_analysis_{profile_type}.txt"
    
    try:
        with open(analysis_file, 'w') as f:
            f.write(f"Benchmark: {benchmark_name}\n")
            f.write(f"Profile Type: {profile_type}\n")
            f.write("=" * 80 + "\n\n")
            f.write(analysis)
        print(f"Analysis saved to: {analysis_file}")
    except Exception as e:
        print(f"Error saving analysis to {analysis_file}: {e}")

def get_default_prompt(template_name: str) -> str:
    """Get the default prompt from the templates directory.
    
    Args:
        template_name: Name of the template file (e.g., 'general_analyze_prompt.txt')
        
    Returns:
        The default prompt text
    """
    template_path = script_dir / "templates" / template_name
    try:
        if template_path.exists():
            with open(template_path, 'r') as f:
                return f.read().strip()
        print(f"Template file not found: {template_path}")
        return ""
    except Exception as e:
        print(f"Error reading template file {template_path}: {e}")
        return ""

def get_general_analyze_prompt(config: Config) -> str:
    """Get the general analysis prompt from configured location.
    
    Args:
        config: The Config instance containing model configuration
        
    Returns:
        The general analysis prompt to use
        
    Raises:
        ValueError: If prompt location is not provided or file cannot be read
    """
    if not config.model_config.general_analyze_prompt_location:
        raise ValueError("general_analyze_prompt_location must be provided in config")
    
    prompt_path = Path(config.model_config.general_analyze_prompt_location)
    if not prompt_path.exists():
        raise ValueError(f"General analyze prompt file not found at: {prompt_path}")
        
    with open(prompt_path, 'r') as f:
        return f.read().strip()

def send_to_model(tag: str, benchmark_name: str, profile_type: str) -> None:
    """Send benchmark profile data to the AI model and save the analysis.
    
    Args:
        tag: The benchmark tag
        benchmark_name: The name of the benchmark
        profile_type: The type of profile
    """
    files = get_benchmark_files(tag, benchmark_name, profile_type)
    
    # Get configuration
    config = ConfigManager.load()
    
    # Get general analysis prompt
    general_prompt = get_general_analyze_prompt(config)
    print("Using general analysis prompt from configuration or default")

    user_prompt = f"""Benchmark: {benchmark_name}
Profile Type: {profile_type}

{files['text_content']}"""

    messages = [
        {"role": "system", "content": general_prompt},
        {"role": "user", "content": user_prompt}
    ]
    
    try:
        client = ConfigManager.get_client()
        response = client.chat.completions.create(
            model=config.model_config.model,
            messages=messages,
            max_tokens=config.model_config.max_tokens,
            temperature=config.model_config.temperature,
            top_p=config.model_config.top_p
        )
        
        analysis = response.choices[0].message.content
        
        # Save the analysis to file
        save_analysis(tag, benchmark_name, profile_type, analysis)
        
    except Exception as e:
        print(f"Error sending data to model for {benchmark_name} ({profile_type}): {e}")

def analyze_all_profiles(tag: str, benchmark_names: List[str], profile_types: List[str]) -> None:
    """Analyze all profiles for given benchmarks and profile types.
    
    Args:
        tag: The benchmark tag (e.g., "test1")
        benchmark_names: List of benchmark names to analyze
        profile_types: List of profile types to analyze (e.g., ["cpu", "memory", "mutex"])
    """
    print(f"\nStarting comprehensive analysis for tag: {tag}")
    print(f"Benchmarks: {', '.join(benchmark_names)}")
    print(f"Profile types: {', '.join(profile_types)}")
    print("=" * 100)
    
    for benchmark in benchmark_names:
        for profile_type in profile_types:
            print(f"\nAnalyzing {benchmark} ({profile_type})...")
            try:
                send_to_model(tag, benchmark, profile_type)
            except Exception as e:
                print(f"Error analyzing {benchmark} ({profile_type}): {e}")
                continue

def analyze_prof_output(tag: str) -> None:
    """Analyze all profiles from a prof run using the tag.
    
    Args:
        tag: The benchmark tag to analyze
    """
    base_dir = Path("bench") / tag
    if not base_dir.exists():
        print(f"Error: No benchmark data found for tag '{tag}'")
        return
        
    # Get all benchmark names from the text directory
    text_dir = base_dir / "text"
    if not text_dir.exists():
        print(f"Error: No text profiles found in {text_dir}")
        return
        
    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        print(f"Error: No benchmark directories found in {text_dir}")
        return
        
    # Get all profile types from the first benchmark's files
    first_benchmark = benchmark_names[0]
    profile_files = list((text_dir / first_benchmark).glob("*.txt"))
    profile_types = [f.stem.split('_')[1] for f in profile_files if '_' in f.stem]
    
    if not profile_types:
        print(f"Error: No profile files found for benchmark {first_benchmark}")
        return
        
    print(f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types")
    analyze_all_profiles(tag, benchmark_names, profile_types)

def get_deep_analysis_files(tag: str, benchmark_name: str) -> Dict[str, str]:
    """Get all text files for a specific benchmark from function directories and text directory.
    
    Args:
        tag: The benchmark tag (e.g., "test1")
        benchmark_name: The name of the benchmark (e.g., "BenchmarkGenPool")
        
    Returns:
        A dictionary containing:
        - functions_content: Combined content of all function profile files from all profile types
        - text_content: Combined content of all text profile files
    """
    base_dir = Path("bench") / tag
    
    # Get all possible function directory types
    function_dirs = [d for d in base_dir.iterdir() if d.is_dir() and d.name.endswith('_functions')]
    
    # Collect content from all function directories
    functions_content = []
    for func_dir in function_dirs:
        profile_type = func_dir.name.replace('_functions', '')
        benchmark_dir = func_dir / benchmark_name
        if benchmark_dir.exists():
            profile_content = []
            for txt_file in benchmark_dir.glob("*.txt"):
                try:
                    with open(txt_file, 'r') as f:
                        # Skip empty files or files with only whitespace
                        content = f.read().strip()
                        if content:
                            # Only include the filename without extension
                            profile_content.append(f"{txt_file.stem}:{content}")
                except Exception as e:
                    print(f"Error reading functions file {txt_file}: {e}")
            if profile_content:
                functions_content.append(f"{profile_type}:{'|'.join(profile_content)}")
    
    # Get files from text directory
    text_dir = base_dir / "text" / benchmark_name
    text_content = []
    if text_dir.exists():
        for txt_file in text_dir.glob("*.txt"):
            try:
                with open(txt_file, 'r') as f:
                    content = f.read().strip()
                    if content:
                        # Only include the filename without extension
                        text_content.append(f"{txt_file.stem}:{content}")
            except Exception as e:
                print(f"Error reading text file {txt_file}: {e}")
    
    return {
        "functions_content": "\n".join(functions_content),
        "text_content": "\n".join(text_content)
    }

def save_deep_analysis(tag: str, benchmark_name: str, analysis: str) -> None:
    """Save the deep AI analysis to a file.
    
    Args:
        tag: The benchmark tag
        benchmark_name: The name of the benchmark
        analysis: The analysis text to save
    """
    # Create the directory structure
    analysis_dir = Path("bench") / tag / "AI" / "deep" / benchmark_name
    analysis_dir.mkdir(parents=True, exist_ok=True)
    
    # Create the analysis file
    analysis_file = analysis_dir / "deep_analysis.txt"
    
    try:
        with open(analysis_file, 'w') as f:
            f.write(f"Benchmark: {benchmark_name}\n")
            f.write("Deep Analysis Report\n")
            f.write("=" * 80 + "\n\n")
            f.write(analysis)
        print(f"Deep analysis saved to: {analysis_file}")
    except Exception as e:
        print(f"Error saving deep analysis to {analysis_file}: {e}")

def get_deep_analysis_prompt(config: Config) -> str:
    """Get the deep analysis prompt from configured location.
    
    Args:
        config: The Config instance containing model configuration
        
    Returns:
        The deep analysis prompt to use
        
    Raises:
        ValueError: If prompt location is not provided or file cannot be read
    """
    if not config.model_config.deep_analyze_prompt_location:
        raise ValueError("deep_analyze_prompt_location must be provided in config")
    
    prompt_path = Path(config.model_config.deep_analyze_prompt_location)
    if not prompt_path.exists():
        raise ValueError(f"Deep analysis prompt file not found at: {prompt_path}")
        
    with open(prompt_path, 'r') as f:
        return f.read().strip()

def send_to_model_deep(tag: str, benchmark_name: str) -> None:
    """Send all benchmark profile data to the AI model for deep analysis.
    
    Args:
        tag: The benchmark tag
        benchmark_name: The name of the benchmark
    """
    files = get_deep_analysis_files(tag, benchmark_name)
    
    # Log the files being analyzed
    print(f"\n{'='*50}\nStarting deep analysis for {benchmark_name} (tag: {tag})")
    
    # Log function profiles
    if files['functions_content']:
        print("\nFunction Profiles being analyzed:")
        for line in files['functions_content'].split('\n'):
            if line.startswith('=== ') or line.startswith('--- '):
                print(line)
    else:
        print("No function profiles found")
    
    # Log text profiles
    if files['text_content']:
        print("\nText Profiles being analyzed:")
        for line in files['text_content'].split('\n'):
            if line.startswith('=== '):
                print(line)
    else:
        print("No text profiles found")
    
    # Get configuration
    config = ConfigManager.load()
    
    # Get deep analysis prompt
    deep_analysis_prompt = get_deep_analysis_prompt(config)
    print("Using deep analysis prompt from configuration or default")

    user_prompt = f"""Benchmark: {benchmark_name}

Function Profiles:
{files['functions_content'].strip()}

Profile Data:
{files['text_content'].strip()}"""

    messages = [
        {"role": "system", "content": deep_analysis_prompt},
        {"role": "user", "content": user_prompt}
    ]
    
    try:
        client = ConfigManager.get_client()
        print(f"\nSending request to model: {config.model_config.model}")
        response = client.chat.completions.create(
            model=config.model_config.model,
            messages=messages,
            max_tokens=config.model_config.max_tokens,
            temperature=config.model_config.temperature,
            top_p=config.model_config.top_p
        )
        
        analysis = response.choices[0].message.content
        print("Successfully received model response")
        
        # Save the analysis to file
        save_deep_analysis(tag, benchmark_name, analysis)
        
    except Exception as e:
        print(f"Error sending data to model for deep analysis of {benchmark_name}: {e}")
        raise

def analyze_all_deep(tag: str, benchmark_names: List[str]) -> None:
    """Perform deep analysis on all given benchmarks.
    
    Args:
        tag: The benchmark tag (e.g., "test1")
        benchmark_names: List of benchmark names to analyze
    """
    print(f"\nStarting deep analysis for tag: {tag}")
    print(f"Benchmarks: {', '.join(benchmark_names)}")
    print("=" * 100)
    
    for benchmark in benchmark_names:
        print(f"\nPerforming deep analysis of {benchmark}...")
        try:
            send_to_model_deep(tag, benchmark)
        except Exception as e:
            print(f"Error in deep analysis of {benchmark}: {e}")
            continue

def analyze_prof_output_deep(tag: str) -> None:
    """Perform deep analysis of all benchmarks from a prof run using the tag.
    
    Args:
        tag: The benchmark tag to analyze
    """
    base_dir = Path("bench") / tag
    
    if not base_dir.exists():
        print(f"Error: No benchmark data found for tag '{tag}'")
        return
        
    # Get all benchmark names from the text directory
    text_dir = base_dir / "text"
    if not text_dir.exists():
        print(f"Error: No text profiles found in {text_dir}")
        return
        
    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        print(f"Error: No benchmark directories found in {text_dir}")
        return
        
    print(f"Found {len(benchmark_names)} benchmarks for deep analysis")
    analyze_all_deep(tag, benchmark_names)

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Analyze benchmark profiles using AI")
    parser.add_argument("tag", help="The benchmark tag to analyze")
    parser.add_argument("-general_analyze", action="store_true", help="Perform general analysis of benchmark profiles")
    parser.add_argument("-deep_analyze", action="store_true", help="Perform deep analysis of benchmark profiles")
    
    args = parser.parse_args()
    if args.general_analyze:
        analyze_prof_output(args.tag)
    elif args.deep_analyze:
        analyze_prof_output_deep(args.tag)
    else:
        print("Please use either -general_analyze or -deep_analyze flag to perform analysis")

# Example usage:
# send_to_model("test1", "BenchmarkGenPool", "cpu")
