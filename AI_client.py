import base64
from pathlib import Path
from typing import Dict, List, Optional
from config_manager import ConfigManager

def read_profile_text_file(file_path: str) -> str:
    """Read and return the contents of a profile text file."""
    try:
        with open(file_path, 'r') as f:
            return f.read()
    except Exception as e:
        print(f"Error reading profile text file {file_path}: {e}")
        return ""

def encode_png_to_base64(file_path: str) -> Optional[str]:
    """Read a PNG file and encode it to base64."""
    try:
        with open(file_path, 'rb') as f:
            return base64.b64encode(f.read()).decode('utf-8')
    except Exception as e:
        print(f"Error reading PNG file {file_path}: {e}")
        return None

def get_benchmark_files(tag: str, benchmark_name: str, profile_type: str) -> Dict[str, str]:
    """Get the text and PNG files for a specific benchmark and profile type.
    
    Args:
        tag: The benchmark tag (e.g., "test1")
        benchmark_name: The name of the benchmark (e.g., "BenchmarkGenPool")
        profile_type: The type of profile (e.g., "cpu", "memory", "mutex")
        
    Returns:
        A dictionary containing:
        - text_content: The content of the profile text file
        - png_base64: The base64 encoded PNG file (if available)
    """
    base_dir = Path("bench") / tag
    
    # Get text file
    text_file = base_dir / "text" / benchmark_name / f"{benchmark_name}_{profile_type}.txt"
    text_content = read_profile_text_file(str(text_file))
    
    # Get PNG file
    png_file = base_dir / f"{profile_type}_functions" / benchmark_name / f"{benchmark_name}_{profile_type}.png"
    png_base64 = encode_png_to_base64(str(png_file)) if png_file.exists() else None
    
    return {
        "text_content": text_content,
        "png_base64": png_base64
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
    
    # Use configured system prompt or default, ignoring the template value
    if (config.model_config.system_prompt 
        and config.model_config.system_prompt != ConfigManager.DEFAULT_SYSTEM_PROMPT_TEMPLATE):
        system_prompt = config.model_config.system_prompt
    else:
        system_prompt = ConfigManager.get_default_system_prompt()

    user_prompt = f"""Benchmark: {benchmark_name}
Profile Type: {profile_type}

Please provide a comprehensive essay-style analysis following the structure above:

{files['text_content']}"""

    # Add PNG if available
    if files['png_base64']:
        user_prompt += f"\n\n[Image data available: {files['png_base64'][:100]}...]"

    messages = [
        {"role": "system", "content": system_prompt},
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

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Analyze benchmark profiles using AI")
    parser.add_argument("tag", help="The benchmark tag to analyze")
    
    args = parser.parse_args()
    analyze_prof_output(args.tag)

# Example usage:
# send_to_model("test1", "BenchmarkGenPool", "cpu")
