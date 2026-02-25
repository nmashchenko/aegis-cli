package db

func (db *DB) seedPapers() error {
	papers := []struct {
		tier, title, url, highlight string
	}{
		// annoyed tier (3+ urges) — dopamine & reward
		{"annoyed", "Dopamine, Reward Prediction, and the Brain",
			"https://pubmed.ncbi.nlm.nih.gov/25234614/",
			"Dopamine neurons fire not for the reward itself, but for the prediction error — the gap between what you expected and what you got."},
		{"annoyed", "Dopamine, Reward Prediction, and the Brain",
			"https://pubmed.ncbi.nlm.nih.gov/25234614/",
			"Unexpected notifications trigger dopamine surges identical to those seen in gambling — your brain is literally playing a slot machine."},
		{"annoyed", "Delay Discounting and Impulsivity",
			"https://pubmed.ncbi.nlm.nih.gov/28212843/",
			"People who check their phones more frequently show steeper delay discounting — they devalue future rewards more heavily."},
		{"annoyed", "Delay Discounting and Impulsivity",
			"https://pubmed.ncbi.nlm.nih.gov/28212843/",
			"The urge to switch tasks is your brain choosing a small immediate reward over a larger delayed one."},
		{"annoyed", "Instant Gratification and Self-Control",
			"https://pubmed.ncbi.nlm.nih.gov/29016274/",
			"Each time you resist an impulse, you strengthen the prefrontal circuits that make the next resistance easier."},
		{"annoyed", "Instant Gratification and Self-Control",
			"https://pubmed.ncbi.nlm.nih.gov/29016274/",
			"The discomfort you feel right now is temporary — it peaks within 3-5 minutes and then fades naturally."},

		// angry tier (5+ urges) — attention & cognition
		{"angry", "Brain Drain: The Mere Presence of One's Own Smartphone Reduces Available Cognitive Capacity",
			"https://pubmed.ncbi.nlm.nih.gov/28493254/",
			"Heavy smartphone use reduces available cognitive capacity even when the phone is face-down and silent."},
		{"angry", "Brain Drain: The Mere Presence of One's Own Smartphone Reduces Available Cognitive Capacity",
			"https://pubmed.ncbi.nlm.nih.gov/28493254/",
			"Participants who left their phone in another room significantly outperformed those with it on the desk — even turned off."},
		{"angry", "The Attentional Cost of Receiving a Cell Phone Notification",
			"https://pubmed.ncbi.nlm.nih.gov/26121498/",
			"A single phone notification — even if you don't check it — causes errors comparable to actually using the phone."},
		{"angry", "The Attentional Cost of Receiving a Cell Phone Notification",
			"https://pubmed.ncbi.nlm.nih.gov/26121498/",
			"It takes an average of 23 minutes to fully regain focus after a single interruption."},
		{"angry", "Media Multitasking and Cognitive Control",
			"https://pubmed.ncbi.nlm.nih.gov/19713392/",
			"Chronic media multitaskers perform worse on task-switching tests — the more you practice distraction, the worse you get at focusing."},

		// chaos tier (10+ urges) — structural brain changes & mental health
		{"chaos", "Microstructure Abnormalities in Adolescents with Internet Addiction Disorder",
			"https://pubmed.ncbi.nlm.nih.gov/21764150/",
			"Brain imaging shows reduced grey matter volume in the prefrontal cortex of people with internet addiction — the area responsible for willpower and decision-making."},
		{"chaos", "Microstructure Abnormalities in Adolescents with Internet Addiction Disorder",
			"https://pubmed.ncbi.nlm.nih.gov/21764150/",
			"White matter integrity in the brain degrades with excessive internet use, impairing communication between brain regions."},
		{"chaos", "Digital Addiction and Sleep, Anxiety, Depression",
			"https://pubmed.ncbi.nlm.nih.gov/31549311/",
			"Excessive screen use is associated with higher rates of anxiety and depression, independent of content consumed."},
		{"chaos", "Digital Addiction and Sleep, Anxiety, Depression",
			"https://pubmed.ncbi.nlm.nih.gov/31549311/",
			"Compulsive digital behavior follows the same neurological pathways as substance addiction — tolerance, withdrawal, and relapse."},
		{"chaos", "Problematic Internet Use and Mental Health",
			"https://pubmed.ncbi.nlm.nih.gov/25405785/",
			"Every hour of unfocused browsing correlates with measurably lower life satisfaction scores across multiple studies."},
	}

	for _, p := range papers {
		_, err := db.conn.Exec(
			"INSERT INTO papers (mood_tier, title, url, highlight) VALUES (?, ?, ?, ?)",
			p.tier, p.title, p.url, p.highlight,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
