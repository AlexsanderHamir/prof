from pathlib import Path
import sys
from typing import Dict, List

from config_manager import Config, ConfigManager
from exit_codes import BENCHMARK_DIRECTORY_MISSING, EXIT_CODE_UNEXPECTED_ERROR, MISSING_PROMPT, MODEL_ANALYSIS_ERROR, PROFILE_READ_EMPTY, PROFILE_READ_ERROR, PROFILE_SAVE_ERROR, TEXT_DIR_EMPTY, TEXT_DIR_MISSING


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
    try:
        response = client.chat.completions.create(model=config.model_config.model, messages=messages, max_tokens=config.model_config.max_tokens, temperature=config.model_config.temperature, top_p=config.model_config.top_p)
        content = response.choices[0].message.content
        if content is None:
            print("No content received from model", file=sys.stderr)
            sys.exit(MODEL_ANALYSIS_ERROR)
        return content
    except Exception as e:
        print(f"Error during model analysis request: {e}", file=sys.stderr)
        sys.exit(MODEL_ANALYSIS_ERROR)


def validate_benchmark_directories(tag: str) -> List[str]:
    base_dir = Path("bench") / tag

    if not base_dir.exists():
        print(f"No benchmark data found for tag '{tag}'", file=sys.stderr)
        sys.exit(BENCHMARK_DIRECTORY_MISSING)

    text_dir = base_dir / "text"
    if not text_dir.exists():
        print(f"No text profiles found in {text_dir}", file=sys.stderr)
        sys.exit(TEXT_DIR_MISSING)

    benchmark_names = [d.name for d in text_dir.iterdir() if d.is_dir()]
    if not benchmark_names:
        print(f"No benchmark directories found in {text_dir}", file=sys.stderr)
        sys.exit(TEXT_DIR_EMPTY)

    return benchmark_names


def read_profile_file(file_path: Path) -> str:
    try:
        content = file_path.read_text().strip()
    except OSError as e:
        print(f"Cannot read profile file {file_path}: {e}", file=sys.stderr)
        sys.exit(PROFILE_READ_ERROR)
    if not content:
        print(f"Profile file {file_path} is empty", file=sys.stderr)
        sys.exit(PROFILE_READ_EMPTY)

    return content


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
    except OSError as e:
        print(f"Cannot read profile file {file_path}: {e}", file=sys.stderr)
        sys.exit(PROFILE_READ_ERROR)


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
    except OSError as e:
        print(f"Cannot save analysis to {analysis_file}: {e}", file=sys.stderr)
        sys.exit(PROFILE_SAVE_ERROR)


def get_file_path(tag: str, benchmark_name: str, profile_type: str) -> Path:
    if ConfigManager.is_flagging:
        return Path("bench") / tag / "text" / benchmark_name / f"{benchmark_name}_{profile_type}.txt"

    analysis_dir = Path("bench") / tag / "AI" / "generalistic" / benchmark_name
    analysis_dir.mkdir(parents=True, exist_ok=True)

    analysis_file = analysis_dir / f"generalistic_analysis_{profile_type}.txt"

    return analysis_file


def get_user_prompt(config: Config) -> str:
    if not config.model_config.prompt_location:
        print("prompt_location must be provided in config", file=sys.stderr)
        sys.exit(MISSING_PROMPT)

    prompt_path = Path(config.model_config.prompt_location)
    if not prompt_path.exists():
        print(f"General analyze prompt file not found at: {prompt_path}", file=sys.stderr)
        sys.exit(MISSING_PROMPT)

    with open(prompt_path, 'r') as f:
        return f.read().strip()


def _format_profile_info(benchmark_name: str, profile_type: str, profile_content: str) -> str:
    return (f"Benchmark: {benchmark_name}\n"
            f"Profile Type: {profile_type}\n\n"
            f"Profile Content: {profile_content}")


def send_to_model(tag: str, benchmark_name: str, profile_type: str) -> None:
    context = f"{benchmark_name} ({profile_type})"
    try:
        profile_data = get_benchmark_file(tag, benchmark_name, profile_type)
        profile_content = profile_data.get('text_content', '')
        if not profile_content:
            print(f"No content found for {context}", file=sys.stderr)
            sys.exit(PROFILE_READ_ERROR)

        config = ConfigManager.get_config()
        user_prompt = get_user_prompt(config)
        profile_info = _format_profile_info(benchmark_name, profile_type, profile_content)
        messages = [
            {
                "role": "system",
                "content": user_prompt
            },
            {
                "role": "user",
                "content": profile_info
            },
        ]
        analysis = request_model_analysis(messages, config)
        save_analysis(tag, benchmark_name, profile_type, analysis)
        print(f"Successfully analyzed and saved results for {context}")

    except Exception as e:
        print(f"Unexpected error for {context}: {e}")
        sys.exit(EXIT_CODE_UNEXPECTED_ERROR)


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
