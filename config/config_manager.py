from pathlib import Path
import sys
from typing import Optional
from openai import OpenAI

from exit_codes import EXIT_CODE_UNEXPECTED_ERROR, MISSING_CONFIG_FILE
from config.helpers import (
    Config,
    validate_config_structure,
    validate_model_config,
    validate_benchmark_configs,
    create_config_template,
    save_template_to_file,
    load_config_from_file,
    create_config_from_data,
    print_template_creation_info,
)


class ConfigManager:
    _config: Optional[Config] = None
    _config_path: Optional[Path] = None
    is_flagging: bool = False

    @staticmethod
    def create_template(output_path: Optional[str] = None) -> None:
        template = create_config_template()
        template_path = Path(output_path) if output_path else Path.cwd() / "config_template.json"
        template_path.parent.mkdir(parents=True, exist_ok=True)

        save_template_to_file(template, template_path)
        ConfigManager._config_path = template_path

        print_template_creation_info(template_path)

    @classmethod
    def setup_from_file(cls, config_path: Path) -> None:
        try:
            config_data = load_config_from_file(config_path)

            validate_config_structure(config_data)

            if "benchmark_configs" not in config_data:
                config_data["benchmark_configs"] = {}

            validate_model_config(config_data["model_config"])

            validate_benchmark_configs(config_data["benchmark_configs"])

            cls._config_path = config_path
            cls._config = create_config_from_data(config_data)
        except Exception as e:
            print(f"Unexpected error during configuration setup: {e}", file=sys.stderr)
            sys.exit(EXIT_CODE_UNEXPECTED_ERROR)

    @classmethod
    def get_config(cls) -> Config:
        if not cls._config:
            print("Configuration file not found. Please ensure the configuration file exists and the name is correct. (config_template.json)", file=sys.stderr)
            sys.exit(MISSING_CONFIG_FILE)

        return cls._config

    @classmethod
    def get_client(cls) -> OpenAI:
        config = cls.get_config()
        return OpenAI(api_key=config.api_key, base_url=config.base_url)
