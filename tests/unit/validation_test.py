import pytest
from config.helpers import validate_ai_config
from exit_codes import CONFIG_VALIDATION_ERROR
from tests.unit.constants import AI_CONFIG_REQUIRED, ALL_BENCHMARKS_AND_PROFILES_TRUE, NO_SPECIFIC_BENCHMARKS, NO_SPECIFIC_PROFILES, UNIVERSAL_PROFILE_FILTER_DICT, UNIVERSAL_PROFILE_FILTER_MISSING_PROFILE_VALUES, UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_INVALID_VALUES, UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_MISSING_FIELDS, UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_NOT_DICT


def test_empty_config(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({})

    captured = capsys.readouterr()
    assert AI_CONFIG_REQUIRED in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_no_specific_benchmarks(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"all_benchmarks": False})

    captured = capsys.readouterr()
    assert NO_SPECIFIC_BENCHMARKS in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_no_specific_profiles(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"all_profiles": False})

    captured = capsys.readouterr()
    assert NO_SPECIFIC_PROFILES in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_all_benchmarks_and_profiles_true(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"all_benchmarks": True, "all_profiles": True, "specific_benchmarks": ["BenchmarkName"], "specific_profiles": ["cpu"]})

    captured = capsys.readouterr()
    assert ALL_BENCHMARKS_AND_PROFILES_TRUE in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_universal_profile_filter_dict(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"universal_profile_filter": "not a dict"})

    captured = capsys.readouterr()
    assert UNIVERSAL_PROFILE_FILTER_DICT in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_universal_profile_filter_missing_profile_values(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"universal_profile_filter": {"not_profile_values": {"flat": 0.0, "flat%": 0.0, "sum%": 0.0, "cum": 0.0, "cum%": 0.0}}})

    captured = capsys.readouterr()
    assert UNIVERSAL_PROFILE_FILTER_MISSING_PROFILE_VALUES in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_universal_profile_filter_dict_profile_values_not_dict(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"universal_profile_filter": {"profile_values": "not a dict"}})

    captured = capsys.readouterr()
    assert UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_NOT_DICT in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_universal_profile_filter_profile_values_missing_fields(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"universal_profile_filter": {"profile_values": {"flat%": 0.0, "sum%": 0.0, "cum": 0.0, "cum%": 0.0}}})

    captured = capsys.readouterr()
    assert UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_MISSING_FIELDS in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR


def test_universal_profile_filter_profile_values_invalid_values(capsys):
    with pytest.raises(SystemExit) as exc_info:
        validate_ai_config({"universal_profile_filter": {"profile_values": {"flat": "not a number", "flat%": 0.0, "sum%": 0.0, "cum": 0.0, "cum%": 0.0}}})

    captured = capsys.readouterr()
    assert UNIVERSAL_PROFILE_FILTER_PROFILE_VALUES_INVALID_VALUES in captured.err
    assert exc_info.value.code == CONFIG_VALIDATION_ERROR
