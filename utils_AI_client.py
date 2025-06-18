from pathlib import Path
from typing import Dict, List

from config_manager import Config, ConfigManager


def log_profile_content(content: str, profile_type: str) -> None:
    if not content:
        print(f"No {profile_type.lower()} profiles found")
        return

    print(f"\n{profile_type} Profiles being analyzed:")
    for line in content.split('\n'):
        if line.startswith('=== ') or line.startswith('--- '):
            print(line)


def request_model_analysis(messages: List[Dict[str, str]], config: Config) -> str:
    client = ConfigManager.get_client()
    print(f"\nSending request to model: {config.model_config.model}")

    response = client.chat.completions.create(model=config.model_config.model, messages=messages, max_tokens=config.model_config.max_tokens, temperature=config.model_config.temperature, top_p=config.model_config.top_p)

    content = response.choices[0].message.content
    if content is None:
        raise ValueError("No content received from model")
    return content


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


def collect_profile_content(directory: Path, file_pattern: str = "*.txt") -> List[str]:
    if not directory.exists():
        return []

    content_list = []
    for txt_file in directory.glob(file_pattern):
        content = read_profile_file(txt_file)
        if content:
            content_list.append(f"{txt_file.stem}:{content}")
    return content_list


def get_function_directories(base_dir: Path) -> List[Path]:
    return [d for d in base_dir.iterdir() if d.is_dir() and d.name.endswith('_functions')]


def collect_function_profiles(base_dir: Path, benchmark_name: str) -> List[str]:
    functions_content = []
    for func_dir in get_function_directories(base_dir):
        profile_type = func_dir.name.replace('_functions', '')
        benchmark_dir = func_dir / benchmark_name

        profile_content = collect_profile_content(benchmark_dir)
        if profile_content:
            functions_content.append(f"{profile_type}:{'|'.join(profile_content)}")

    return functions_content


def collect_text_profiles(base_dir: Path, benchmark_name: str) -> List[str]:
    text_dir = base_dir / "text" / benchmark_name
    return collect_profile_content(text_dir)


def read_profile_text_file(file_path: str) -> str:
    try:
        with open(file_path, 'r') as f:
            return f.read().strip()
    except Exception as e:
        raise ProfileReadError(f"Error reading profile file {file_path}: {e}")


def get_benchmark_file(tag: str, benchmark_name: str, profile_type: str) -> Dict[str, str]:
    base_dir = Path("bench") / tag
    text_dir = base_dir / "text" / benchmark_name
    profile_file = text_dir / f"{benchmark_name}_{profile_type}.txt"

    return {
        "text_content": read_profile_text_file(str(profile_file)),
    }


def save_analysis(tag: str, benchmark_name: str, profile_type: str, analysis: str) -> None:

    analysis_file = get_file_path(tag, benchmark_name, profile_type)

    try:
        with open(analysis_file, 'w') as f:
            f.write(f"Benchmark: {benchmark_name}\n")
            f.write(f"Profile Type: {profile_type}\n")
            f.write("=" * 80 + "\n\n")
            f.write(analysis)
        print(f"Analysis saved to: {analysis_file}")
    except Exception as e:
        print(f"Error saving analysis to {analysis_file}: {e}")


def get_file_path(tag: str, benchmark_name: str, profile_type: str) -> Path:
    if ConfigManager.is_flagging:
        return Path("bench") / tag / "text" / benchmark_name / f"{benchmark_name}_{profile_type}.txt"

    # Create the directory structure
    analysis_dir = Path("bench") / tag / "AI" / "generalistic" / benchmark_name
    analysis_dir.mkdir(parents=True, exist_ok=True)

    # Create the analysis file
    analysis_file = analysis_dir / f"generalistic_analysis_{profile_type}.txt"

    return analysis_file


def get_default_prompt(template_name: str) -> str:
    """Get a default prompt template if none is provided in config."""
    template_dir = Path(__file__).parent / "prompts"
    template_file = template_dir / f"{template_name}.txt"

    if not template_file.exists():
        raise ValueError(f"Default prompt template not found: {template_file}")

    with open(template_file, 'r') as f:
        return f.read().strip()


def get_general_analyze_prompt(config: Config) -> str:
    if not config.model_config.prompt_location:
        raise ValueError("prompt_location must be provided in config")

    prompt_path = Path(config.model_config.prompt_location)
    if not prompt_path.exists():
        raise ValueError(f"General analyze prompt file not found at: {prompt_path}")

    with open(prompt_path, 'r') as f:
        return f.read().strip()


def _build_profile_info(benchmark_name: str, profile_type: str, profile_content: str) -> str:
    return (f"Benchmark: {benchmark_name}\n"
            f"Profile Type: {profile_type}\n\n"
            f"Profile Content: {profile_content}")


def _print_and_raise_error(context: str, error: Exception) -> None:
    print(f"Error analyzing {context}: {error}")
    raise error


def send_to_model(tag: str, benchmark_name: str, profile_type: str) -> None:
    context = f"{benchmark_name} ({profile_type})"
    try:
        profile_data = get_benchmark_file(tag, benchmark_name, profile_type)
        profile_content = profile_data.get('text_content', '')
        if not profile_content:
            raise ValueError(f"No content found for {context}")

        config = ConfigManager.load()
        general_prompt = get_general_analyze_prompt(config)
        profile_info = _build_profile_info(benchmark_name, profile_type, profile_content)
        messages = [
            {
                "role": "system",
                "content": general_prompt
            },
            {
                "role": "user",
                "content": profile_info
            },
        ]
        analysis = request_model_analysis(messages, config)
        save_analysis(tag, benchmark_name, profile_type, analysis)
        print(f"Successfully analyzed and saved results for {context}")

    except ValueError as e:
        print(f"Validation error for {context}: {e}")
        raise
    except Exception as e:
        _print_and_raise_error(context, e)


def analyze_all_profiles(tag: str, benchmark_names: List[str], profile_types: List[str]) -> None:
    print(f"\nStarting comprehensive analysis for tag: {tag}")
    print(f"Benchmarks: {', '.join(benchmark_names)}")
    print(f"Profile types: {', '.join(profile_types)}")
    print("=" * 100)

    profile_types = [profile_type for profile_type in profile_types if "trace" not in profile_type]

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
    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        raise ValueError(f"No benchmark directories found in {text_dir}")
    return benchmark_names


def get_profile_types(text_dir: Path, benchmark_name: str) -> List[str]:
    benchmark_dir = text_dir / benchmark_name
    if not benchmark_dir.exists():
        raise ValueError(f"No benchmark directory found: {benchmark_dir}")

    profile_files = list(benchmark_dir.glob(f"{benchmark_name}_*.txt"))
    if not profile_files:
        raise ValueError(f"No profile files found for {benchmark_name}")

    return [f.stem.split('_', 1)[1] for f in profile_files]
