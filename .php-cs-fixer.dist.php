<?php

$finder = PhpCsFixer\Finder::create()
    ->in(__DIR__.'/backend-php')
    ->exclude('include/lang')
    ->name('*.php');

return (new PhpCsFixer\Config())
    ->setRiskyAllowed(true)
    ->setRules([
        '@PSR12' => true,
        'array_syntax' => ['syntax' => 'short'],
        'no_unused_imports' => true,
        'ordered_imports' => ['sort_algorithm' => 'none'],
        'no_empty_statement' => true,
        'no_extra_blank_lines' => true,
        'no_trailing_whitespace' => true,
        'no_trailing_whitespace_in_comment' => true,
        'single_blank_line_at_eof' => true,
        'single_quote' => true,
        'standardize_not_equals' => true,
        'ternary_operator_spaces' => true,
        'uniform_variable_name' => true,
        'concat_space' => ['spacing' => 'one'],
    ])
    ->setFinder($finder);
