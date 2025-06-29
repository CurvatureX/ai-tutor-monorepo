#!/usr/bin/env python3
"""
Example Usage of the Prompt System for AI English Learning Service

This file demonstrates how to use the new prompt.py module to create
dynamic, context-aware prompts for different learning scenarios.
"""

import asyncio
import httpx
from prompt import (
    create_system_prompt,
    create_conversation_starter,
    create_roleplay_prompt,
    create_assessment_prompt,
    get_encouragement_phrases,
    ConversationContext,
    LearningLevel,
)

# Configuration
BASE_URL = "http://localhost:8000"
TEST_TOKEN = "demo-token"


def example_basic_prompts():
    """Demonstrate basic prompt creation"""
    print("🎯 Basic Prompt Examples")
    print("=" * 50)

    # Basic general prompt
    general_prompt = create_system_prompt()
    print("General Prompt:")
    print(general_prompt[:200] + "...\n")

    # Grammar-focused prompt for beginners
    grammar_prompt = create_system_prompt(
        context=ConversationContext.GRAMMAR, level=LearningLevel.BEGINNER
    )
    print("Grammar Prompt (Beginner):")
    print(grammar_prompt[:200] + "...\n")

    # Business English for advanced learners
    business_prompt = create_system_prompt(
        context=ConversationContext.BUSINESS_ENGLISH,
        level=LearningLevel.ADVANCED,
        topic="job interviews",
    )
    print("Business English Prompt (Advanced):")
    print(business_prompt[:200] + "...\n")


def example_personalized_prompts():
    """Demonstrate personalized prompt creation"""
    print("👤 Personalized Prompt Examples")
    print("=" * 50)

    # Prompt with learning goals and interests
    personalized_prompt = create_system_prompt(
        context=ConversationContext.CONVERSATION_PRACTICE,
        level=LearningLevel.INTERMEDIATE,
        topic="travel",
        learning_goals=[
            "Improve speaking fluency",
            "Learn travel vocabulary",
            "Practice past tense storytelling",
        ],
        user_interests=["photography", "hiking", "cultural experiences"],
    )
    print("Personalized Travel Prompt:")
    print(personalized_prompt)
    print()


def example_conversation_starters():
    """Demonstrate conversation starter generation"""
    print("💬 Conversation Starter Examples")
    print("=" * 50)

    contexts = [
        (ConversationContext.GENERAL, LearningLevel.BEGINNER, None),
        (ConversationContext.VOCABULARY, LearningLevel.INTERMEDIATE, "technology"),
        (ConversationContext.BUSINESS_ENGLISH, LearningLevel.ADVANCED, "networking"),
    ]

    for context, level, topic in contexts:
        starter = create_conversation_starter(context, level, topic)
        print(f"{context.value.title()} ({level.value}) - Topic: {topic}")
        print(f"Starter: {starter}")
        print()


def example_roleplay_prompts():
    """Demonstrate roleplay scenario prompts"""
    print("🎭 Roleplay Prompt Examples")
    print("=" * 50)

    scenarios = [
        ("restaurant", "customer", "waiter"),
        ("job interview", "candidate", "interviewer"),
        ("hotel check-in", "guest", "receptionist"),
        ("doctor's office", "patient", "doctor"),
    ]

    for scenario, user_role, ai_role in scenarios:
        roleplay_prompt = create_roleplay_prompt(scenario, user_role, ai_role)
        print(f"Scenario: {scenario.title()}")
        print(f"Prompt: {roleplay_prompt[:150]}...")
        print()


def example_assessment_prompts():
    """Demonstrate assessment prompt creation"""
    print("📊 Assessment Prompt Examples")
    print("=" * 50)

    skill_areas = ["speaking", "grammar", "vocabulary", "comprehension"]

    for skill in skill_areas:
        assessment_prompt = create_assessment_prompt(skill)
        print(f"{skill.title()} Assessment:")
        print(assessment_prompt[:200] + "...")
        print()


def example_encouragement_phrases():
    """Demonstrate level-appropriate encouragement"""
    print("🌟 Encouragement Phrase Examples")
    print("=" * 50)

    for level in LearningLevel:
        phrases = get_encouragement_phrases(level)
        print(f"{level.value.title()} Level:")
        print(f"  {', '.join(phrases[:3])}...")
        print()


async def test_api_with_enhanced_features():
    """Test the API endpoints with enhanced prompt features"""
    print("🔌 API Testing with Enhanced Features")
    print("=" * 50)

    headers = {"Authorization": f"Bearer {TEST_TOKEN}"}

    async with httpx.AsyncClient() as client:
        try:
            # Test conversation starter endpoint
            response = await client.get(
                f"{BASE_URL}/v1/conversation/starter",
                headers=headers,
                params={
                    "context": "vocabulary",
                    "level": "intermediate",
                    "topic": "food",
                },
            )
            if response.status_code == 200:
                result = response.json()
                print("✅ Conversation Starter:")
                print(f"   Context: {result['context']}")
                print(f"   Level: {result['level']}")
                print(f"   Topic: {result['topic']}")
                print(f"   Starter: {result['starter']}")
                print()

            # Test enhanced session creation
            response = await client.post(
                f"{BASE_URL}/v1/language/session",
                headers=headers,
                params={
                    "language": "English",
                    "level": "intermediate",
                    "context": "business_english",
                    "topic": "presentations",
                    "goals": "improve presentation skills, learn business vocabulary",
                    "interests": "technology, entrepreneurship",
                },
            )
            if response.status_code == 200:
                session = response.json()
                print("✅ Enhanced Session Created:")
                print(f"   ID: {session['id']}")
                print(f"   Context: {session['context']}")
                print(f"   Topic: {session['topic']}")
                print(f"   Goals: {session['goals']}")
                print(f"   Interests: {session['interests']}")
                print()

            # Test chat completion with context awareness
            response = await client.post(
                f"{BASE_URL}/v1/chat/completions",
                headers=headers,
                json={
                    "model": "gpt-4-tutor",
                    "messages": [
                        {
                            "role": "user",
                            "content": "I want to practice grammar, specifically past tense",
                        }
                    ],
                    "stream": False,
                },
            )
            if response.status_code == 200:
                result = response.json()
                content = result["choices"][0]["message"]["content"]
                print("✅ Context-Aware Response:")
                print(f"   {content[:200]}...")
                print()

        except Exception as e:
            print(f"❌ API Test Error: {e}")


def main():
    """Run all examples"""
    print("🎓 AI English Learning Service - Prompt System Examples")
    print("=" * 70)
    print()

    # Run examples
    example_basic_prompts()
    example_personalized_prompts()
    example_conversation_starters()
    example_roleplay_prompts()
    example_assessment_prompts()
    example_encouragement_phrases()

    # Test API if service is running
    print("🔧 To test the API endpoints, make sure the service is running:")
    print("   python app.py")
    print()
    print("Then run the API tests:")
    print("   python example_prompt_usage.py --test-api")
    print()


async def main_async():
    """Run async examples and API tests"""
    main()
    await test_api_with_enhanced_features()


if __name__ == "__main__":
    import sys

    if "--test-api" in sys.argv:
        print("🧪 Testing API with enhanced features...")
        asyncio.run(main_async())
    else:
        main()
