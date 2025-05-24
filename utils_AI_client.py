import os
from pathlib import Path
from typing import Dict, List

from config_manager import Config, ConfigManager


def save_deep_analysis(tag: str, benchmark_name: str, analysis: str) -> None:
    analysis_dir = Path("bench") / tag / "AI" / "deep" / benchmark_name
    analysis_dir.mkdir(parents=True, exist_ok=True)

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
    if not config.model_config.deep_analyze_prompt_location:
        raise ValueError(
            "deep_analyze_prompt_location must be provided in config")

    prompt_path = Path(config.model_config.deep_analyze_prompt_location)
    if not prompt_path.exists():
        raise ValueError(
            f"Deep analysis prompt file not found at: {prompt_path}")

    with open(prompt_path, 'r') as f:
        return f.read().strip()


def log_profile_content(content: str, profile_type: str) -> None:
    if not content:
        print(f"No {profile_type.lower()} profiles found")
        return

    print(f"\n{profile_type} Profiles being analyzed:")
    for line in content.split('\n'):
        if line.startswith('=== ') or line.startswith('--- '):
            print(line)


def prepare_deep_analysis_prompt(
        benchmark_name: str, files: Dict[str, str]) -> List[Dict[str, str]]:
    """Prepare a structured prompt for deep analysis of benchmark profiles.
    
    This function constructs a prompt for AI analysis by:
    1. Loading the deep analysis prompt template from config
    2. Combining benchmark name with function and text profiles
    3. Structuring the prompt in a format suitable for the AI model
    
    Args:
        benchmark_name: Name of the benchmark to analyze
        files: Dictionary containing function and text profile contents
        
    Returns:
        List of message dictionaries formatted for the AI model API
        
    Raises:
        ValueError: If deep analysis prompt template is not configured
    """
    config = ConfigManager.load()
    deep_analysis_prompt = get_deep_analysis_prompt(config)

    user_prompt = f"""Benchmark: {benchmark_name}

Function Profiles:
{files['functions_content'].strip()}

Profile Data:
{files['text_content'].strip()}"""

    return [{
        "role": "system",
        "content": deep_analysis_prompt
    }, {
        "role": "user",
        "content": user_prompt
    }]


def request_model_analysis(messages: List[Dict[str, str]],
                           config: Config) -> str:
    """Send a request to the AI model for profile analysis.
    
    This function handles the interaction with the AI model API:
    1. Configures the model client with appropriate settings
    2. Sends the prepared messages to the model
    3. Processes and returns the model's response
    
    Args:
        messages: List of message dictionaries for the model
        config: Configuration containing model settings
        
    Returns:
        String containing the model's analysis
        
    Raises:
        RuntimeError: If the model request fails
    """
    client = ConfigManager.get_client()
    print(f"\nSending request to model: {config.model_config.model}")

    response = client.chat.completions.create(
        model=config.model_config.model,
        messages=messages,
        max_tokens=config.model_config.max_tokens,
        temperature=config.model_config.temperature,
        top_p=config.model_config.top_p)

    return response.choices[0].message.content


def validate_benchmark_directories(tag: str) -> list[str]:
    base_dir = Path("bench") / tag

    if not base_dir.exists():
        raise ValueError(f"No benchmark data found for tag '{tag}'")

    text_dir = base_dir / "text"
    if not text_dir.exists():
        raise ValueError(f"No text profiles found in {text_dir}")

    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        raise ValueError(f"No benchmark directories found in {text_dir}")

    return benchmark_names


class ProfileReadError(Exception):
    """Custom exception for profile file reading errors."""
    pass


def read_profile_file(file_path: Path) -> str:
    try:
        with open(file_path, 'r') as f:
            content = f.read().strip()
            if not content:
                raise ProfileReadError(f"Profile file {file_path} is empty")
            return content
    except FileNotFoundError:
        raise ProfileReadError(f"Profile file not found: {file_path}")
    except Exception as e:
        raise ProfileReadError(f"Error reading profile file {file_path}: {e}")


def collect_profile_content(directory: Path,
                            file_pattern: str = "*.txt") -> List[str]:
    if not directory.exists():
        return []

    content_list = []
    for txt_file in directory.glob(file_pattern):
        content = read_profile_file(txt_file)
        if content:
            content_list.append(f"{txt_file.stem}:{content}")
    return content_list


def get_function_directories(base_dir: Path) -> List[Path]:
    return [
        d for d in base_dir.iterdir()
        if d.is_dir() and d.name.endswith('_functions')
    ]


def collect_function_profiles(base_dir: Path,
                              benchmark_name: str) -> List[str]:
    """Collect and organize function profiles from multiple profile types.
    
    This function aggregates function profiles across different profile types:
    1. Discovers all function profile directories
    2. Collects profiles for each benchmark
    3. Organizes profiles by type and content
    4. Handles file reading and error cases
    
    Args:
        base_dir: Base directory containing profile data
        benchmark_name: Name of the benchmark to collect profiles for
        
    Returns:
        List of formatted profile strings, organized by profile type
    """
    functions_content = []
    for func_dir in get_function_directories(base_dir):
        profile_type = func_dir.name.replace('_functions', '')
        benchmark_dir = func_dir / benchmark_name

        profile_content = collect_profile_content(benchmark_dir)
        if profile_content:
            functions_content.append(
                f"{profile_type}:{'|'.join(profile_content)}")

    return functions_content


def collect_text_profiles(base_dir: Path, benchmark_name: str) -> List[str]:
    text_dir = base_dir / "text" / benchmark_name
    return collect_profile_content(text_dir)


def send_to_model_deep(tag: str, benchmark_name: str) -> None:
    """Perform deep analysis of benchmark profiles using AI model.
    
    This function orchestrates the entire deep analysis process:
    1. Gathers all relevant profile data
    2. Prepares the analysis prompt
    3. Sends data to the AI model
    4. Saves the analysis results
    
    The function handles the complete workflow from data collection
    to result storage, including error handling and logging.
    
    Args:
        tag: Unique identifier for this analysis run
        benchmark_name: Name of the benchmark to analyze
        
    Raises:
        ProfileReadError: If profile files cannot be read
        RuntimeError: If model analysis fails
    """
    # Step 1: Gather profile data
    files = get_deep_analysis_files(tag, benchmark_name)

    # Step 2: Log what we're analyzing
    print(
        f"\n{'='*50}\nStarting deep analysis for {benchmark_name} (tag: {tag})"
    )
    log_profile_content(files['functions_content'], "Function")
    log_profile_content(files['text_content'], "Text")

    # Step 3: Prepare the model request
    config = ConfigManager.load()
    print("Using deep analysis prompt from configuration or default")
    messages = prepare_deep_analysis_prompt(benchmark_name, files)

    try:
        # Step 4: Get model analysis
        analysis = request_model_analysis(messages, config)
        print("Successfully received model response")

        # Step 5: Save the results
        save_deep_analysis(tag, benchmark_name, analysis)

    except Exception as e:
        print(
            f"Error sending data to model for deep analysis of {benchmark_name}: {e}"
        )
        raise


def analyze_all_deep(tag: str, benchmark_names: List[str]) -> None:
    """Perform deep analysis for multiple benchmarks.
    
    This function manages the analysis of multiple benchmarks:
    1. Validates the benchmark directory structure
    2. Processes each benchmark sequentially
    3. Handles errors for individual benchmarks
    4. Provides progress feedback
    
    Args:
        tag: Unique identifier for this analysis run
        benchmark_names: List of benchmarks to analyze
        
    Note:
        Errors in individual benchmark analysis are logged but don't
        stop the overall process
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


def get_deep_analysis_files(tag: str, benchmark_name: str) -> Dict[str, str]:
    """Collect and organize all files needed for deep analysis.
    
    This function gathers and structures all necessary profile data:
    1. Collects function profiles from all profile types
    2. Gathers text profiles for the benchmark
    3. Organizes the data into a structured format
    
    Args:
        tag: Unique identifier for this analysis run
        benchmark_name: Name of the benchmark to analyze
        
    Returns:
        Dictionary containing:
        - functions_content: Combined function profiles
        - text_content: Combined text profiles
    """
    base_dir = Path("bench") / tag

    functions_content = collect_function_profiles(base_dir, benchmark_name)
    text_content = collect_text_profiles(base_dir, benchmark_name)

    return {
        "functions_content": "\n".join(functions_content),
        "text_content": "\n".join(text_content)
    }


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


def get_benchmark_files(tag: str, benchmark_name: str,
                        profile_type: str) -> Dict[str, str]:
    base_dir = Path("bench") / tag

    # Get text file
    text_file = base_dir / "text" / benchmark_name / f"{benchmark_name}_{profile_type}.txt"
    text_content = read_profile_text_file(str(text_file))

    return {"text_content": text_content}


def save_analysis(tag: str, benchmark_name: str, profile_type: str,
                  analysis: str) -> None:
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
    if not config.model_config.general_analyze_prompt_location:
        raise ValueError(
            "general_analyze_prompt_location must be provided in config")

    prompt_path = Path(config.model_config.general_analyze_prompt_location)
    if not prompt_path.exists():
        raise ValueError(
            f"General analyze prompt file not found at: {prompt_path}")

    with open(prompt_path, 'r') as f:
        return f.read().strip()


def request_model_general(messages: List[Dict[str, str]],
                          config: Config) -> str:
    client = ConfigManager.get_client()
    try:
        response = client.chat.completions.create(
            model=config.model_config.model,
            messages=messages,
            max_tokens=config.model_config.max_tokens,
            temperature=config.model_config.temperature,
            top_p=config.model_config.top_p)
        return response.choices[0].message.content
    except Exception as e:
        raise Exception(f"Model request failed: {str(e)}")


def send_to_model(tag: str, benchmark_name: str, profile_type: str) -> None:
    try:
        files = get_benchmark_files(tag, benchmark_name, profile_type)
        if not files['text_content']:
            raise ValueError(
                f"No content found for {benchmark_name} ({profile_type})")

        config = ConfigManager.load()
        general_prompt = get_general_analyze_prompt(config)

        user_prompt = f"""Benchmark: {benchmark_name}
Profile Type: {profile_type}

{files['text_content']}"""

        messages = [{
            "role": "system",
            "content": general_prompt
        }, {
            "role": "user",
            "content": user_prompt
        }]

        print(f"Analyzing {benchmark_name} ({profile_type})...")
        analysis = request_model_general(messages, config)
        save_analysis(tag, benchmark_name, profile_type, analysis)
        print(
            f"Successfully analyzed and saved results for {benchmark_name} ({profile_type})"
        )

    except ValueError as e:
        print(f"Validation error for {benchmark_name} ({profile_type}): {e}")
        raise
    except Exception as e:
        print(f"Error analyzing {benchmark_name} ({profile_type}): {e}")
        raise


def analyze_all_profiles(tag: str, benchmark_names: List[str],
                         profile_types: List[str]) -> None:
    print(f"\nStarting comprehensive analysis for tag: {tag}")
    print(f"Benchmarks: {', '.join(benchmark_names)}")
    print(f"Profile types: {', '.join(profile_types)}")
    print("=" * 100)

    for benchmark in benchmark_names:
        for profile_type in profile_types:
            print(f"\nAnalyzing {benchmark} ({profile_type})...")
            send_to_model(tag, benchmark, profile_type)


def validate_benchmark_directory(tag: str) -> Path:
    base_dir = Path("bench") / tag
    if not base_dir.exists():
        raise ValueError(f"No benchmark data found for tag '{tag}'")
    return base_dir


def get_benchmark_names(text_dir: Path) -> List[str]:
    if not text_dir.exists():
        raise ValueError(f"No text profiles found in {text_dir}")

    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        raise ValueError(f"No benchmark directories found in {text_dir}")

    return benchmark_names


def get_profile_types(text_dir: Path, benchmark_name: str) -> List[str]:
    profile_files = list((text_dir / benchmark_name).glob("*.txt"))
    profile_types = [
        f.stem.split('_')[1] for f in profile_files if '_' in f.stem
    ]

    if not profile_types:
        raise ValueError(
            f"No profile files found for benchmark {benchmark_name}")

    return profile_types
