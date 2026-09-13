import 'package:flutter/material.dart';
import 'features/auth/screens/login_screen.dart';
import 'features/channels/screens/channels_screen.dart';
import 'features/voice/screens/voice_screen.dart';
import 'features/profile/screens/profile_screen.dart';

class RadioPXApp extends StatelessWidget {
  const RadioPXApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'RadioPX',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2196F3),
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: const MainScreen(),
    );
  }
}

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _currentIndex = 0;

  final List<Widget> _screens = [
    const ChannelsScreen(),
    const VoiceScreen(),
    const ProfileScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _screens[_currentIndex],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.radio),
            selectedIcon: Icon(Icons.radio, color: Color(0xFF2196F3)),
            label: 'Canais',
          ),
          NavigationDestination(
            icon: Icon(Icons.mic),
            selectedIcon: Icon(Icons.mic, color: Color(0xFF2196F3)),
            label: 'Falar',
          ),
          NavigationDestination(
            icon: Icon(Icons.person),
            selectedIcon: Icon(Icons.person, color: Color(0xFF2196F3)),
            label: 'Perfil',
          ),
        ],
      ),
    );
  }
}
