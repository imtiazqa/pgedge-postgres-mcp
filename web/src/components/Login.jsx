/*-------------------------------------------------------------------------
 *
 * pgEdge MCP Client - Login Page
 *
 * Portions copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 * Styled to match pgEdge Cloud product aesthetics
 *
 *-------------------------------------------------------------------------
 */

import React, { useState, useEffect } from 'react';
import {
    Box,
    Card,
    CardContent,
    TextField,
    Button,
    Typography,
    Alert,
    Container,
    keyframes,
    alpha,
} from '@mui/material';
import { useAuth } from '../contexts/AuthContext';
import logoLight from '../assets/images/logo-light.png';

// Subtle floating animation for decorative elements
const float = keyframes`
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-15px); }
`;

const pulse = keyframes`
  0%, 100% { opacity: 0.4; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.05); }
`;

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [warning, setWarning] = useState('');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();

    // Check for disconnect message on mount
    useEffect(() => {
        const disconnectMsg = sessionStorage.getItem('disconnectMessage');
        if (disconnectMsg) {
            setWarning(disconnectMsg);
            sessionStorage.removeItem('disconnectMessage');
        }
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setWarning('');
        setLoading(true);

        try {
            await login(username, password);
        } catch (err) {
            setError(err.message || 'Failed to login');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Box
            sx={{
                minHeight: '100vh',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                position: 'relative',
                overflow: 'hidden',
                // pgEdge Cloud background
                backgroundImage: 'url(https://a.storyblok.com/f/187930/1200x560/7852cd29b7/home-page-hero-bg-1200.jpg)',
                backgroundSize: 'cover',
                backgroundPosition: 'center',
                backgroundRepeat: 'no-repeat',
                '&::before': {
                    content: '""',
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.85) 0%, rgba(30, 41, 59, 0.75) 100%)',
                    zIndex: 0,
                },
            }}
        >
            {/* Decorative geometric shapes */}
            <Box
                sx={{
                    position: 'absolute',
                    top: '10%',
                    left: '5%',
                    width: 200,
                    height: 200,
                    borderRadius: '50%',
                    background: `linear-gradient(135deg, ${alpha('#15AABF', 0.2)} 0%, ${alpha('#22B8CF', 0.1)} 100%)`,
                    filter: 'blur(60px)',
                    animation: `${pulse} 8s ease-in-out infinite`,
                    zIndex: 1,
                }}
            />
            <Box
                sx={{
                    position: 'absolute',
                    bottom: '15%',
                    right: '10%',
                    width: 250,
                    height: 250,
                    borderRadius: '50%',
                    background: `linear-gradient(135deg, ${alpha('#3B82F6', 0.15)} 0%, ${alpha('#15AABF', 0.1)} 100%)`,
                    filter: 'blur(80px)',
                    animation: `${pulse} 10s ease-in-out infinite`,
                    animationDelay: '2s',
                    zIndex: 1,
                }}
            />

            {/* Subtle grid pattern overlay */}
            <Box
                sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    backgroundImage: `
                        linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
                        linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px)
                    `,
                    backgroundSize: '50px 50px',
                    zIndex: 1,
                }}
            />

            {/* Floating decorative elements */}
            <Box
                sx={{
                    position: 'absolute',
                    top: '20%',
                    right: '20%',
                    width: 60,
                    height: 60,
                    border: `2px solid ${alpha('#15AABF', 0.3)}`,
                    borderRadius: '12px',
                    transform: 'rotate(15deg)',
                    animation: `${float} 6s ease-in-out infinite`,
                    zIndex: 1,
                }}
            />
            <Box
                sx={{
                    position: 'absolute',
                    bottom: '25%',
                    left: '15%',
                    width: 40,
                    height: 40,
                    border: `2px solid ${alpha('#22B8CF', 0.25)}`,
                    borderRadius: '50%',
                    animation: `${float} 8s ease-in-out infinite`,
                    animationDelay: '1s',
                    zIndex: 1,
                }}
            />
            <Box
                sx={{
                    position: 'absolute',
                    top: '60%',
                    right: '8%',
                    width: 30,
                    height: 30,
                    backgroundColor: alpha('#15AABF', 0.2),
                    borderRadius: '6px',
                    transform: 'rotate(45deg)',
                    animation: `${float} 7s ease-in-out infinite`,
                    animationDelay: '3s',
                    zIndex: 1,
                }}
            />

            <Container maxWidth="sm" sx={{ position: 'relative', zIndex: 2 }}>
                <Card
                    elevation={24}
                    sx={{
                        backdropFilter: 'blur(20px)',
                        backgroundColor: 'rgba(255, 255, 255, 0.98)',
                        borderRadius: 3,
                        border: '1px solid rgba(255, 255, 255, 0.2)',
                        overflow: 'visible',
                        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
                    }}
                >
                    <CardContent sx={{ p: { xs: 3, sm: 5 } }}>
                        <Box sx={{ textAlign: 'center', mb: 4 }}>
                            <Box
                                component="img"
                                src={logoLight}
                                alt="pgEdge"
                                sx={{
                                    height: '48px',
                                    mb: 2,
                                }}
                            />
                            <Typography
                                variant="h5"
                                component="h1"
                                sx={{
                                    fontWeight: 600,
                                    color: '#1F2937',
                                    mb: 0.5,
                                }}
                            >
                                Natural Language Agent
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#6B7280' }}>
                                Sign in to continue
                            </Typography>
                        </Box>

                        {warning && (
                            <Alert
                                severity="warning"
                                sx={{
                                    mb: 3,
                                    borderRadius: 1,
                                    '& .MuiAlert-icon': {
                                        color: '#F59E0B',
                                    },
                                }}
                                onClose={() => setWarning('')}
                            >
                                {warning}
                            </Alert>
                        )}

                        {error && (
                            <Alert
                                severity="error"
                                sx={{
                                    mb: 3,
                                    borderRadius: 1,
                                }}
                            >
                                {error}
                            </Alert>
                        )}

                        <form onSubmit={handleSubmit} noValidate>
                            <TextField
                                fullWidth
                                label="Username"
                                type="text"
                                name="username"
                                id="username"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                margin="normal"
                                required
                                autoFocus
                                disabled={loading}
                                inputProps={{
                                    autoComplete: 'off',
                                }}
                                sx={{
                                    '& .MuiOutlinedInput-root': {
                                        borderRadius: 1,
                                        '&:hover .MuiOutlinedInput-notchedOutline': {
                                            borderColor: '#9CA3AF',
                                        },
                                        '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
                                            borderColor: '#15AABF',
                                            borderWidth: 2,
                                        },
                                    },
                                    '& .MuiInputLabel-root.Mui-focused': {
                                        color: '#15AABF',
                                    },
                                }}
                            />

                            <TextField
                                fullWidth
                                label="Password"
                                type="password"
                                name="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                margin="normal"
                                required
                                disabled={loading}
                                inputProps={{
                                    autoComplete: 'current-password',
                                }}
                                sx={{
                                    '& .MuiOutlinedInput-root': {
                                        borderRadius: 1,
                                        '&:hover .MuiOutlinedInput-notchedOutline': {
                                            borderColor: '#9CA3AF',
                                        },
                                        '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
                                            borderColor: '#15AABF',
                                            borderWidth: 2,
                                        },
                                    },
                                    '& .MuiInputLabel-root.Mui-focused': {
                                        color: '#15AABF',
                                    },
                                }}
                            />

                            <Button
                                fullWidth
                                type="submit"
                                variant="contained"
                                size="large"
                                disabled={loading}
                                sx={{
                                    mt: 3,
                                    py: 1.5,
                                    borderRadius: 1,
                                    fontWeight: 600,
                                    fontSize: '1rem',
                                    textTransform: 'none',
                                    background: '#15AABF',
                                    boxShadow: '0 4px 14px 0 rgba(14, 165, 233, 0.39)',
                                    '&:hover': {
                                        background: '#0C8599',
                                        boxShadow: '0 6px 20px 0 rgba(14, 165, 233, 0.5)',
                                    },
                                    '&.Mui-disabled': {
                                        background: '#E5E7EB',
                                        color: '#9CA3AF',
                                    },
                                }}
                            >
                                {loading ? 'Signing in...' : 'Sign In'}
                            </Button>
                        </form>

                        <Box sx={{ mt: 3, textAlign: 'center' }}>
                            <Typography variant="caption" sx={{ color: '#9CA3AF' }}>
                                Contact your administrator to create an account
                            </Typography>
                        </Box>
                    </CardContent>
                </Card>

                {/* Copyright footer */}
                <Typography
                    variant="caption"
                    sx={{
                        display: 'block',
                        textAlign: 'center',
                        mt: 3,
                        color: 'rgba(255, 255, 255, 0.6)',
                    }}
                >
                    &copy; 2025 pgEdge, Inc.
                </Typography>
            </Container>
        </Box>
    );
};

export default Login;
