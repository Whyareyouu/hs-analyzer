import type {
	OverlayDeckCard,
	OpponentPlayedCard,
} from '@/shared/model/overlay';

type InspectableCard = OverlayDeckCard | OpponentPlayedCard;

export function CardTooltip({ card }: { card: InspectableCard | null }) {
	if (!card) {
		return null;
	}

	return (
		<div className='h-[350px] w-[225px] overflow-hidden'>
			{card.image ? (
				<img
					src={card.image}
					alt={card.name}
					className='h-full w-full object-contain'
					draggable={false}
				/>
			) : (
				<div
					className='h-full w-full bg-[radial-gradient(circle_at_top,rgba(148,163,184,0.24),rgba(15,23,42,0.96))]'
					aria-hidden='true'
				/>
			)}
		</div>
	);
}
