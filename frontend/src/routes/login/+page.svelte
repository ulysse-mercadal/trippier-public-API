<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth';
	import { register, login, verifyCode, getMe, ApiError } from '$lib/api';
	import TopoCanvas from '$lib/components/TopoCanvas.svelte';

	type Stage = 'form' | 'otp';
	type Mode  = 'login' | 'register';

	let mode: Mode   = 'register';
	let stage: Stage = 'form';
	let email        = '';
	let password     = '';
	let error        = '';
	let loading      = false;

	// OTP — 6 individual digit boxes
	let digits: string[] = Array(6).fill('');
	let digitEls: HTMLInputElement[] = [];

	onMount(async () => {
		const stored = auth.getStoredToken();
		if (!stored) return;
		try {
			const user = await getMe(stored);
			auth.init(stored, user);
			goto('/dashboard');
		} catch {
			// token expired — stay on login
		}
	});

	async function submitForm() {
		error   = '';
		loading = true;
		try {
			if (mode === 'register') {
				await register(email, password);
				stage = 'otp';
				digits = Array(6).fill('');
				setTimeout(() => digitEls[0]?.focus(), 50);
			} else {
				const token = await login(email, password);
				const user  = await getMe(token);
				auth.init(token, user);
				auth.storeToken(token);
				goto('/dashboard');
			}
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong';
		} finally {
			loading = false;
		}
	}

	async function submitOTP() {
		error   = '';
		const code = digits.join('');
		if (code.length !== 6) { error = 'Enter all 6 digits'; return; }
		loading = true;
		try {
			const token = await verifyCode(email, code);
			const user  = await getMe(token);
			auth.init(token, user);
			auth.storeToken(token);
			goto('/dashboard');
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong';
			digits = Array(6).fill('');
			setTimeout(() => digitEls[0]?.focus(), 50);
		} finally {
			loading = false;
		}
	}

	function handleDigitInput(e: Event, i: number) {
		const input = e.target as HTMLInputElement;
		const val   = input.value.replace(/\D/g, '').slice(-1);
		digits[i]   = val;
		if (val && i < 5) digitEls[i + 1]?.focus();
	}

	function handleDigitKeydown(e: KeyboardEvent, i: number) {
		if (e.key === 'Backspace' && !digits[i] && i > 0) {
			digits[i - 1] = '';
			digitEls[i - 1]?.focus();
		}
		if (e.key === 'ArrowLeft'  && i > 0) digitEls[i - 1]?.focus();
		if (e.key === 'ArrowRight' && i < 5) digitEls[i + 1]?.focus();
	}

	function handleDigitPaste(e: ClipboardEvent) {
		e.preventDefault();
		const pasted = (e.clipboardData?.getData('text') ?? '').replace(/\D/g, '').slice(0, 6);
		digits = [...pasted.split(''), ...Array(6).fill('')].slice(0, 6);
		const next = Math.min(pasted.length, 5);
		digitEls[next]?.focus();
	}

	function switchMode(m: Mode) {
		mode  = m;
		error = '';
		if (stage === 'otp') stage = 'form';
	}
</script>

<svelte:head>
	<title>Sign in · Trippier API</title>
</svelte:head>

<div class="split">
	<!-- ── Left panel: form ── -->
	<div class="left">
		<a href="/" class="logo">tripp<em>ier</em></a>

		<div class="form-wrap">
			{#if stage === 'form'}
				<div class="tabs">
					<button class="tab" class:active={mode === 'register'} on:click={() => switchMode('register')}>Create account</button>
					<button class="tab" class:active={mode === 'login'}    on:click={() => switchMode('login')}>Sign in</button>
				</div>

				<div class="form-head">
					<h1>{mode === 'register' ? 'Start for free' : 'Welcome back'}</h1>
					<p class="form-sub">
						{mode === 'register'
							? '1 000 free tokens. Your API key in 30 seconds.'
							: 'Sign in to manage your API keys.'}
					</p>
				</div>

				{#if error}<p class="alert alert-error">{error}</p>{/if}

				<form on:submit|preventDefault={submitForm}>
					<div class="field">
						<label for="email">Email</label>
						<input id="email" type="email" bind:value={email} placeholder="you@example.com" required autocomplete="email" />
					</div>
					<div class="field">
						<label for="password">Password</label>
						<input id="password" type="password" bind:value={password} placeholder="••••••••" required autocomplete={mode === 'login' ? 'current-password' : 'new-password'} />
					</div>
					<button class="btn btn-primary submit-btn" type="submit" disabled={loading}>
						{#if loading}…{:else if mode === 'register'}Create account & get key{:else}Sign in{/if}
					</button>
				</form>

			{:else}
				<!-- OTP stage -->
				<button class="btn-link" on:click={() => { stage = 'form'; error = ''; }}>← Back</button>

				<div class="form-head">
					<h1>Check your inbox</h1>
					<p class="form-sub">
						We sent a 6-digit code to <strong>{email}</strong>. Enter it below — the code expires in 15 minutes.
					</p>
				</div>

				{#if error}<p class="alert alert-error">{error}</p>{/if}

				<form on:submit|preventDefault={submitOTP}>
					<div class="otp-row">
						{#each digits as _, i}
							<input
								type="text"
								inputmode="numeric"
								maxlength="1"
								class="otp-digit"
								value={digits[i]}
								bind:this={digitEls[i]}
								on:input={(e) => handleDigitInput(e, i)}
								on:keydown={(e) => handleDigitKeydown(e, i)}
								on:paste={handleDigitPaste}
								autocomplete="one-time-code"
							/>
						{/each}
					</div>
					<button class="btn btn-primary submit-btn" type="submit" disabled={loading || digits.join('').length < 6}>
						{loading ? '…' : 'Verify & get API key'}
					</button>
				</form>

				<p class="resend-hint">Didn't receive it? Check your spam folder.</p>
			{/if}
		</div>
	</div>

	<!-- ── Right panel: white + gray topo ── -->
	<div class="right">
		{#if browser}<TopoCanvas fixed={false} />{/if}
		<div class="right-overlay">
			<blockquote class="right-quote">
				<p>"Went from idea to working POI search in an afternoon."</p>
				<cite>— a happy builder</cite>
			</blockquote>
		</div>
	</div>
</div>

<style>
	.split {
		display: flex;
		min-height: 100vh;
	}

	/* ── Left ── */
	.left {
		width: 45%;
		min-width: 340px;
		background: #0d0d0d;
		display: flex;
		flex-direction: column;
		padding: 2.5rem 3rem;
		position: relative;
		z-index: 1;
	}

	.logo {
		font-size: 1.15rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		color: var(--text);
		display: block;
		margin-bottom: auto;
	}
	.logo em { font-style: normal; color: var(--accent); }

	.form-wrap {
		width: 100%;
		max-width: 360px;
		margin: auto;
		padding: 2rem 0;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.tabs {
		display: flex;
		background: rgba(0, 0, 0, 0.4);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 3px;
	}

	.tab {
		flex: 1;
		background: none;
		border: none;
		border-radius: calc(var(--radius) - 2px);
		color: var(--muted);
		font-size: 0.85rem;
		font-weight: 500;
		padding: 0.4rem 0;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	.tab.active { background: rgba(255, 255, 255, 0.07); color: var(--text); }

	.form-head { display: flex; flex-direction: column; gap: 0.4rem; }
	.form-head h1 { font-size: 1.4rem; font-weight: 800; letter-spacing: -0.02em; }
	.form-sub { font-size: 0.85rem; color: var(--muted); line-height: 1.5; }

	.submit-btn { width: 100%; justify-content: center; padding: 0.7rem; font-size: 0.9rem; }

	.btn-link {
		background: none;
		border: none;
		color: var(--muted);
		font-size: 0.82rem;
		cursor: pointer;
		padding: 0;
		transition: color 0.15s;
		text-align: left;
	}
	.btn-link:hover { color: var(--text); }

	/* ── OTP ── */
	.otp-row {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.otp-digit {
		flex: 1;
		aspect-ratio: 1;
		text-align: center;
		font-size: 1.4rem;
		font-weight: 700;
		font-family: var(--mono);
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		color: var(--text);
		outline: none;
		transition: border-color 0.15s;
		caret-color: transparent;
	}
	.otp-digit:focus { border-color: var(--accent); }

	.resend-hint {
		font-size: 0.78rem;
		color: var(--muted);
		text-align: center;
		line-height: 1.5;
	}

	/* ── Right panel: light ── */
	.right {
		flex: 1;
		position: relative;
		overflow: hidden;
		background: #f4f4f2;
		display: flex;
		align-items: flex-end;
	}

	.right-overlay {
		position: relative;
		z-index: 2;
		padding: 3rem;
		width: 100%;
	}

	.right-quote {
		max-width: 340px;
		border-left: 2px solid #10b981;
		padding-left: 1.25rem;
	}

	.right-quote p {
		font-size: 1rem;
		color: rgba(30, 30, 30, 0.65);
		line-height: 1.6;
		font-style: italic;
		margin-bottom: 0.5rem;
	}

	.right-quote cite {
		font-size: 0.78rem;
		color: rgba(30, 30, 30, 0.4);
		font-style: normal;
	}

	/* ── Responsive ── */
	@media (max-width: 700px) {
		.split { flex-direction: column; }
		.left  { width: 100%; min-width: 0; padding: 2rem 1.5rem; }
		.right { display: none; }
	}
</style>
