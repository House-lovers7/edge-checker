import { useResultLoader } from './hooks/useResultLoader';
import { useTimelinePlayer } from './hooks/useTimelinePlayer';
import { FileDropZone } from './components/drop/FileDropZone';
import { Header } from './components/layout/Header';
import { MainLayout } from './components/layout/MainLayout';

export default function App() {
  const { result, error, loadFromFile, loadResult, reset } = useResultLoader();
  const totalSeconds = result?.timeline.length ?? 0;
  const player = useTimelinePlayer(totalSeconds);

  if (!result) {
    return (
      <FileDropZone
        onLoad={loadResult}
        onFileLoad={loadFromFile}
        error={error}
      />
    );
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Header result={result} onReset={reset} />
      <MainLayout result={result} player={player} />
    </div>
  );
}
