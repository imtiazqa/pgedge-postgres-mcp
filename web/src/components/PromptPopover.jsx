/*-------------------------------------------------------------------------
 *
 * pgEdge MCP Client - Prompt Popover Component
 *
 * Portions copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import {
    Popover,
    Box,
    Typography,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    TextField,
    Button,
    Alert,
    Divider,
    CircularProgress,
} from '@mui/material';
import {
    PlayArrow as PlayArrowIcon,
} from '@mui/icons-material';

const PromptPopover = ({
    anchorEl,
    open,
    onClose,
    prompts = [],
    onExecute,
    executing = false,
}) => {
    const [selectedPromptName, setSelectedPromptName] = useState('');
    const [argumentValues, setArgumentValues] = useState({});
    const [validationErrors, setValidationErrors] = useState({});

    // Get the selected prompt object
    const selectedPrompt = prompts.find(p => p.name === selectedPromptName) || null;

    // Reset when prompt changes
    useEffect(() => {
        if (selectedPrompt) {
            const initialValues = {};
            if (selectedPrompt.arguments) {
                selectedPrompt.arguments.forEach(arg => {
                    initialValues[arg.name] = '';
                });
            }
            setArgumentValues(initialValues);
            setValidationErrors({});
        }
    }, [selectedPromptName]);

    const handleArgumentChange = (argName, value) => {
        setArgumentValues(prev => ({
            ...prev,
            [argName]: value,
        }));

        // Clear validation error when user starts typing
        if (validationErrors[argName]) {
            setValidationErrors(prev => {
                const newErrors = { ...prev };
                delete newErrors[argName];
                return newErrors;
            });
        }
    };

    const validateArguments = () => {
        const errors = {};
        let isValid = true;

        if (selectedPrompt && selectedPrompt.arguments) {
            selectedPrompt.arguments.forEach(arg => {
                if (arg.required && !argumentValues[arg.name]?.trim()) {
                    errors[arg.name] = 'This field is required';
                    isValid = false;
                }
            });
        }

        setValidationErrors(errors);
        return isValid;
    };

    const handleExecute = () => {
        if (!selectedPrompt) return;

        if (!validateArguments()) {
            return;
        }

        // Only send non-empty arguments
        const filteredArgs = {};
        Object.entries(argumentValues).forEach(([key, value]) => {
            if (value && value.trim()) {
                filteredArgs[key] = value.trim();
            }
        });

        onExecute(selectedPrompt.name, filteredArgs);

        // Reset and close
        setSelectedPromptName('');
        setArgumentValues({});
        setValidationErrors({});
        onClose();
    };

    const handleClose = () => {
        setSelectedPromptName('');
        setArgumentValues({});
        setValidationErrors({});
        onClose();
    };

    const hasArguments = selectedPrompt?.arguments && selectedPrompt.arguments.length > 0;

    return (
        <Popover
            open={open}
            anchorEl={anchorEl}
            onClose={handleClose}
            disableScrollLock
            anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
            }}
            transformOrigin={{
                vertical: 'bottom',
                horizontal: 'right',
            }}
        >
            <Box sx={{ p: 2, minWidth: 400, maxWidth: 500 }}>
                <Typography variant="h6" sx={{ mb: 2 }}>
                    Execute Prompt
                </Typography>

                <Divider sx={{ mb: 2 }} />

                {/* Prompt Selection */}
                <FormControl fullWidth size="small" sx={{ mb: 2 }}>
                    <InputLabel id="prompt-popover-select-label">Select Prompt</InputLabel>
                    <Select
                        labelId="prompt-popover-select-label"
                        id="prompt-popover-select"
                        value={selectedPromptName}
                        label="Select Prompt"
                        onChange={(e) => setSelectedPromptName(e.target.value)}
                        disabled={executing}
                    >
                        <MenuItem value="">
                            <em>Select a prompt...</em>
                        </MenuItem>
                        {[...prompts].sort((a, b) => a.name.localeCompare(b.name)).map((prompt) => (
                            <MenuItem key={prompt.name} value={prompt.name}>
                                {prompt.name}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                {/* Selected Prompt Info */}
                {selectedPrompt && (
                    <>
                        {/* Description */}
                        {selectedPrompt.description && (
                            <Alert severity="info" sx={{ mb: 2 }}>
                                {selectedPrompt.description}
                            </Alert>
                        )}

                        {/* Arguments */}
                        {hasArguments && (
                            <Box sx={{ mb: 2 }}>
                                <Typography variant="subtitle2" gutterBottom>
                                    Arguments:
                                </Typography>
                                {selectedPrompt.arguments.map((arg) => (
                                    <TextField
                                        key={arg.name}
                                        fullWidth
                                        size="small"
                                        label={arg.name}
                                        value={argumentValues[arg.name] || ''}
                                        onChange={(e) => handleArgumentChange(arg.name, e.target.value)}
                                        error={!!validationErrors[arg.name]}
                                        helperText={validationErrors[arg.name] || arg.description}
                                        required={arg.required}
                                        disabled={executing}
                                        sx={{ mb: 2 }}
                                    />
                                ))}
                            </Box>
                        )}

                        {/* Execute Button */}
                        <Button
                            fullWidth
                            variant="contained"
                            onClick={handleExecute}
                            disabled={executing}
                            startIcon={executing ? <CircularProgress size={16} /> : <PlayArrowIcon />}
                        >
                            {executing ? 'Executing...' : 'Execute Prompt'}
                        </Button>
                    </>
                )}

                {/* Help Text when no prompt selected */}
                {!selectedPrompt && (
                    <Typography variant="body2" color="text.secondary">
                        Select a prompt from the dropdown above to get started.
                    </Typography>
                )}
            </Box>
        </Popover>
    );
};

PromptPopover.propTypes = {
    anchorEl: PropTypes.object,
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
    prompts: PropTypes.arrayOf(PropTypes.shape({
        name: PropTypes.string.isRequired,
        description: PropTypes.string,
        arguments: PropTypes.arrayOf(PropTypes.shape({
            name: PropTypes.string.isRequired,
            description: PropTypes.string,
            required: PropTypes.bool,
        })),
    })),
    onExecute: PropTypes.func.isRequired,
    executing: PropTypes.bool,
};

export default PromptPopover;
