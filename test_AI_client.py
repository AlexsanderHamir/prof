import pytest
from AI_client import *

from unittest.mock import patch, MagicMock

def test_analyze_prof_output_general_valid_case():
    with patch('AI_client.validate_benchmark_directories', return_value=['benchmark1', 'benchmark2']) as mock_validate, \
         patch('AI_client.get_profile_types', return_value=['type1', 'type2']) as mock_get_types, \
         patch('AI_client.analyze_all_profiles') as mock_analyze:

        analyze_prof_output_general('test_tag')

        mock_validate.assert_called_once_with('test_tag')
        mock_get_types.assert_called_once()
        mock_analyze.assert_called_once_with('test_tag', ['benchmark1', 'benchmark2'], ['type1', 'type2'])

def test_analyze_prof_output_general_exception_handling():
    with patch('AI_client.validate_benchmark_directories', side_effect=Exception('Test Exception')):
        with pytest.raises(Exception, match='Test Exception'):
            analyze_prof_output_general('test_tag')
