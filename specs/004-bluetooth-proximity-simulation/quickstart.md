# Quickstart: Simulação de Proximidade via Bluetooth

## Execução

```bash
cd frontend
flutter pub get
flutter run
```

## Verificação automatizada

```bash
cd frontend
flutter analyze
flutter test
```

## Verificação do manifesto mesclado

```bash
cd frontend
flutter build apk --debug
```

Conferir o manifesto mesclado gerado em `frontend/build/app/intermediates/merged_manifests/debug/AndroidManifest.xml`:

- presente: `android.permission.BLUETOOTH` com `android:maxSdkVersion="30"`;
- ausentes: `BLUETOOTH_ADMIN`, `BLUETOOTH_SCAN`, `BLUETOOTH_ADVERTISE`, `BLUETOOTH_CONNECT`;
- presentes: permissões pré-existentes de INTERNET, RECORD_AUDIO, ACCESS_FINE_LOCATION, ACCESS_COARSE_LOCATION e MODIFY_AUDIO_SETTINGS.

> Status da entrega (22/11/2026): `flutter pub get` resolveu `bluetooth_manager
> 2.0.1`; `flutter analyze` terminou sem erros (13 infos pré-existentes fora da
> feature); `flutter test` passou com 66 testes; `flutter build apk --debug`
> gerou `build/app/outputs/flutter-apk/app-debug.apk`. O manifesto mesclado foi
> inspecionado e contém `BLUETOOTH` com `maxSdkVersion="30"`, sem
> `BLUETOOTH_ADMIN`, `BLUETOOTH_SCAN`, `BLUETOOTH_ADVERTISE` ou
> `BLUETOOTH_CONNECT`. Falta apenas a validação manual em dispositivo Android.

## Validação manual em Android

1. Abrir o app autenticado com Bluetooth ligado: indicador mostra "Alguém por perto".
2. Desligar o Bluetooth nas configurações do sistema com o app aberto: indicador muda para "Ninguém por perto" em até 3 segundos.
3. Ligar novamente: indicador volta para "Alguém por perto" em até 3 segundos.
4. Forçar erro/estado desconhecido com provider de teste: indicador mostra "Indisponível".
5. Alternar entre Canais, Falar e Perfil: indicador permanece visível e consistente.
6. Confirmar que nenhum diálogo de permissão Nearby Devices aparece e que canais, WebSocket e PTT continuam inalterados.

## Fora de escopo

iOS, Web e Windows: a factory retorna provider indisponível e nenhuma chamada nativa é feita.
