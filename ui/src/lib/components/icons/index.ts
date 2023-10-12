import {
	AlertCircle,
	ArrowUpDown,
	BookPlus,
	CalendarPlus,
	CalendarSearch,
	Check,
	CheckCircle2,
	ChevronLeft,
	ChevronRight,
	Circle,
	CircleOff,
	CornerRightDown,
	CornerUpLeft,
	FileCode2,
	FileText,
	FileVideo,
	Files,
	Hash,
	Hexagon,
	Info,
	MoreHorizontal,
	Play,
	RefreshCw,
	Search,
	Trash2,
	X,
	XCircle,
	type Icon as LucideIcon
} from 'lucide-svelte';
import Github from './Github.svelte';
import HalfCircle from './HalfCircle.svelte';
import Moon from './Moon.svelte';
import Path from './Path.svelte';
import Sun from './Sun.svelte';

export type Icon = LucideIcon;

export const Icons = {
	arrowUpDown: ArrowUpDown,
	bookPlus: BookPlus,
	calendarPlus: CalendarPlus,
	calendarSearch: CalendarSearch,
	check: Check,
	checkCircle: CheckCircle2,
	chevronLeft: ChevronLeft,
	chevronRight: ChevronRight,
	circle: Circle,
	circleOff: CircleOff,
	cornerRightDown: CornerRightDown,
	cornerUpLeft: CornerUpLeft,
	delete: Trash2,
	errorCircle: XCircle,
	fileHtml: FileCode2,
	filePdf: FileText,
	files: Files,
	fileVideo: FileVideo,
	gitHub: Github,
	halfCircle: HalfCircle,
	hash: Hash,
	infoCircle: Info,
	logo: Hexagon,
	moon: Moon,
	moreHorizontal: MoreHorizontal,
	path: Path,
	play: Play,
	refresh: RefreshCw,
	search: Search,
	sun: Sun,
	warningCircle: AlertCircle,
	x: X
};